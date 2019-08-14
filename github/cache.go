package github

import (
	"context"
	"fmt"
	"time"

	githubApi "github.com/google/go-github/github"
	goCache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	diffUsersFlightGroup *singleflight.Group
	changesFlightGroup   *singleflight.Group
	configFlightGroup    *singleflight.Group
	diffUsersMap         *goCache.Cache
	changesMap           *goCache.Cache
	configMap            *goCache.Cache
	client               *githubApi.Client
	configName           string
}

func newCache(configName string, client *githubApi.Client, ttl time.Duration) *Cache {

	return &Cache{
		diffUsersFlightGroup: &singleflight.Group{},
		changesFlightGroup:   &singleflight.Group{},
		configFlightGroup:    &singleflight.Group{},
		diffUsersMap:         goCache.New(ttl, ttl*2),
		changesMap:           goCache.New(ttl, ttl*2),
		configMap:            goCache.New(ttl, ttl*2),
		client:               client,
		configName:           configName,
	}
}

func (cache *Cache) GetChanges(ctx context.Context, diffURL string) ([]string, error) {
	in, ok := cache.changesMap.Get(diffURL)
	if ok {
		return in.([]string), nil
	}
	in, err, _ := cache.changesFlightGroup.Do(diffURL, func() (interface{}, error) {
		changes, err := readChangedFiles(ctx, diffURL)
		if err != nil {
			return nil, fmt.Errorf("unable to read diffs: %s", err.Error())
		}
		cache.changesMap.SetDefault(diffURL, changes)
		return changes, nil
	})
	if err != nil {
		return nil, err
	}
	return in.([]string), nil
}

func (cache *Cache) GetConfig(ctx context.Context, name, ref string) (*Config, error) {
	id := name + "/" + ref
	in, ok := cache.configMap.Get(id)
	if ok {
		return in.(*Config), nil
	}
	in, err, _ := cache.changesFlightGroup.Do(id, func() (interface{}, error) {
		config, err := fetchRemoteConfig(ctx, cache.client, name, ref, cache.configName)
		if err != nil {
			return nil, err
		}
		cache.configMap.SetDefault(id, config)
		return config, nil
	})
	if err != nil {
		return nil, err
	}
	return in.(*Config), nil
}

func (cache *Cache) GetRelatedUsers(ctx context.Context, name, ref, diffURL string) (map[string]bool, error) {
	id := fmt.Sprintf("%s/%s#%s", name, ref, diffURL)
	in, ok := cache.diffUsersMap.Get(id)
	if ok {
		return in.(map[string]bool), nil
	}
	in, err, _ := cache.diffUsersFlightGroup.Do(id, func() (interface{}, error) {
		config, err := cache.GetConfig(ctx, name, ref)
		if err != nil {
			return nil, err
		}

		changes, err := cache.GetChanges(ctx, diffURL)
		if err != nil {
			return nil, err
		}

		users, err := findRelatedUsers(config, changes)
		if err != nil {
			return nil, err
		}
		cache.diffUsersMap.SetDefault(id, users)
		return users, nil
	})

	if err != nil {
		return nil, err
	}
	return in.(map[string]bool), nil
}
