package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Token string `envconfig:"GITHUB_TOKEN" required:"true"`
}

func NewClient(cfg GithubConfig) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	return client
}
