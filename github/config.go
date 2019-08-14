package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	githubApi "github.com/google/go-github/github"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Settings ConfigSettings          `yaml:"settings"`
	Roles    map[string][]ConfigRole `yaml:"roles"`
	Access   map[string][]string     `yaml:"access"`
}

type ConfigSettings struct {
	NumAcceptRequired int    `yaml:"numAcceptRequired"`
	MergeMethod       string `yaml:"mergeMethod"`
}

type ConfigRole struct {
	User string `yaml:"user"`
}

func fetchRemoteConfig(ctx context.Context, client *githubApi.Client, repo, branch, configName string) (*Config, error) {
	names := strings.Split(repo, "/")
	br, _, err := client.Repositories.GetBranch(ctx, names[0], names[1], branch)
	if err != nil {
		return nil, err
	}

	if configName == "" {
		configName = ".rbacbot.yaml"
	}
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, br.Commit.GetSHA(), configName))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(bodyBytes, config)
	return config, err
}
