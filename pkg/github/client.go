package github

import (
	"Aj-vrod/github-notifier/types"
	"context"
	"errors"
	"log"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHubClientInterface defines the interface for GitHub client operations
type GitHubClientInterface interface {
	GetPRState(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error)
}

type GithubConfig struct {
	Token string `envconfig:"GITHUB_TOKEN" required:"true"`
}

type GithubClient struct {
	client *githubv4.Client
}

// Ensure GithubClient implements GitHubClientInterface
var _ GitHubClientInterface = (*GithubClient)(nil)

func NewClient(cfg GithubConfig) *GithubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	return &GithubClient{client: client}
}

func (c *GithubClient) GetPRState(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
	log.Printf("Getting state for PR with reference %s/%s#%d\n", prInfo.Owner, prInfo.Repo, prInfo.Number)
	variables := map[string]interface{}{
		"owner":    githubv4.String(prInfo.Owner),
		"repo":     githubv4.String(prInfo.Repo),
		"prNumber": githubv4.Int(prInfo.Number),
	}
	var query types.PRQuery

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}

	// Check if the PR exists by verifying if the commits field is nil (GitHub returns null for non-existent PRs)
	if query.Repository.PullRequest.Commits.Nodes == nil {
		return types.PRQuery{}, errors.New("PR does not exists")
	}

	log.Println("PR state was successfully found")

	return query, nil
}
