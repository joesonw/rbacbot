package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	githubApi "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githubWebhook "gopkg.in/go-playground/webhooks.v5/github"
)

type Server struct {
	hook   *githubWebhook.Webhook
	client *githubApi.Client
	cache  *Cache
}

func New(configName, secret, token string, ttl time.Duration) (*Server, error) {
	hook, err := githubWebhook.New(githubWebhook.Options.Secret(secret))
	if err != nil {
		return nil, err
	}
	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	client := githubApi.NewClient(tc)
	return &Server{
		hook:   hook,
		client: client,
		cache:  newCache(configName, client, ttl),
	}, nil
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	payload, err := s.hook.Parse(req, githubWebhook.PullRequestEvent, githubWebhook.PullRequestReviewEvent)
	if err != nil {
		if err == githubWebhook.ErrEventNotFound {
			http.NotFound(res, req)
			return
		}
		log.Printf("unable to parse request: %s\n", err.Error())
		http.Error(res, "Internal Error", http.StatusInternalServerError)
		return
	}

	switch payload.(type) {
	case githubWebhook.PullRequestPayload:
		{
			pr := payload.(githubWebhook.PullRequestPayload)
			if strings.EqualFold(pr.Action, "closed") || strings.EqualFold(pr.Action, "locked") {
				break
			}

			config, err := s.cache.GetConfig(req.Context(), pr.Repository.FullName, pr.PullRequest.Base.Ref)
			if err != nil {
				log.Printf("unable to get remote config: %s\n", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}
			relatedUsers, err := s.cache.GetRelatedUsers(req.Context(), pr.Repository.FullName, pr.PullRequest.Base.Ref, pr.PullRequest.DiffURL)

			if err != nil {
				log.Printf("unable to find related users: %s\n", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}
			content := fmt.Sprintf("found following users allowed to merge this request (at least %d of them must \"accept\").\n", config.Settings.NumAcceptRequired)
			for user := range relatedUsers {
				content += fmt.Sprintf("@%s \n", user)
			}
			comment := &githubApi.IssueComment{
				ID:   githubApi.Int64(pr.Number),
				Body: githubApi.String(content),
			}
			_, _, err = s.client.Issues.CreateComment(req.Context(), pr.Repository.Owner.Login, pr.Repository.Name, int(pr.Number), comment)
			if err != nil {
				log.Printf("unable to create comment: %s", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}
		}
	case githubWebhook.PullRequestReviewPayload:
		{
			review := payload.(githubWebhook.PullRequestReviewPayload)
			if !strings.EqualFold(review.Review.State, "approved") {
				break
			}

			config, err := s.cache.GetConfig(req.Context(), review.Repository.FullName, review.PullRequest.Base.Ref)
			if err != nil {
				log.Printf("unable to get remote config: %s\n", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}
			relatedUsers, err := s.cache.GetRelatedUsers(req.Context(), review.Repository.FullName, review.PullRequest.Base.Ref, review.PullRequest.DiffURL)
			if err != nil {
				log.Printf("unable to find related users: %s\n", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}

			if !relatedUsers[review.Sender.Login] {
				break
			}

			reviews, _, err := s.client.PullRequests.ListReviews(req.Context(), review.Repository.Owner.Login, review.Repository.Name, int(review.PullRequest.Number), &githubApi.ListOptions{})
			if err != nil {
				log.Printf("unable to list reviews: %s", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}

			content := fmt.Sprintf("@%s approved, ", review.Review.User.Login)
			count := 0
			for _, r := range reviews {
				if strings.EqualFold(r.GetState(), "approved") && relatedUsers[r.GetUser().GetLogin()] && r.GetCommitID() == review.PullRequest.Head.Sha {
					count++
				}
			}
			content += fmt.Sprintf("%d/%d.", count, config.Settings.NumAcceptRequired)
			if count >= config.Settings.NumAcceptRequired {
				content += " :white_check_mark:"
			}

			comment := &githubApi.IssueComment{
				ID:   githubApi.Int64(review.PullRequest.Number),
				Body: githubApi.String(content),
			}
			_, _, err = s.client.Issues.CreateComment(req.Context(), review.Repository.Owner.Login, review.Repository.Name, int(review.PullRequest.Number), comment)
			if err != nil {
				log.Printf("unable to create comment: %s", err.Error())
				http.Error(res, "Internal Error", http.StatusInternalServerError)
				return
			}
			if count >= config.Settings.NumAcceptRequired {
				method := "merge"
				if config.Settings.MergeMethod != "" {
					method = config.Settings.MergeMethod
				}
				options := &githubApi.PullRequestOptions{
					MergeMethod: method,
				}
				_, _, err = s.client.PullRequests.Merge(req.Context(), review.Repository.Owner.Login, review.Repository.Name, int(review.PullRequest.Number), review.PullRequest.Title, options)
				if err != nil {
					log.Printf("unable to merge PR %d: %s", review.PullRequest.Number, err.Error())
					http.Error(res, "Internal Error", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	res.WriteHeader(200)
	res.Write([]byte("OK"))
}
