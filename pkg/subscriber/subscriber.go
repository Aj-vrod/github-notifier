package subscriber

import (
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/types"
	"context"
	"fmt"
)

type Subscriber struct {
	ghClient *github.GithubClient
}

func NewSubscriber(ghClient *github.GithubClient) *Subscriber {
	return &Subscriber{
		ghClient: ghClient,
	}
}

func (s *Subscriber) CheckPRState(ctx context.Context, prInfo types.PRInfo) (types.PRState, error) {
	ghState, err := s.ghClient.GetPRState(ctx, prInfo.Owner, prInfo.Repo, prInfo.Number)
	if err != nil {
		return types.PRState{}, err
	}

	// Temporally print the GitHub PR state for debugging purposes
	fmt.Println("GitHub PR State:", ghState)

	return types.PRState{
		Exists:   true,
		Comments: "ghState.Repository.PullRequest.Comments.Nodes,",
		Commits:  "ghState.Repository.PullRequest.Commits.Nodes",
	}, nil
}
