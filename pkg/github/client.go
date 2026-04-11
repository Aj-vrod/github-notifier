package github

import (
	"context"
	"log"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Token string `envconfig:"GITHUB_TOKEN" required:"true"`
}

type GithubClient struct {
	client *githubv4.Client
}

func NewClient(cfg GithubConfig) *GithubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	return &GithubClient{client: client}
}

func (c *GithubClient) GetPRState(ctx context.Context, owner, repo string, prNumber int) (any, error) {
	variables := map[string]interface{}{
		"owner":    githubv4.String(owner),
		"repo":     githubv4.String(repo),
		"prNumber": githubv4.Int(prNumber),
	}
	var query struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes []struct {
						Body   string
						Author struct {
							Login string
						}
						CreatedAt time.Time
					}
				} `graphql:"comments(first: 100)"`
				Commits struct {
					Nodes []struct {
						Commit struct {
							Message string
						}
					}
				} `graphql:"commits(first: 100)"`
			} `graphql:"pullRequest(number: $prNumber)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}

	return query, nil
}
