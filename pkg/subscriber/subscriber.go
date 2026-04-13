package subscriber

import (
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/types"
	"context"
	"log"
)

type Subscriber struct {
	ghClient github.GitHubClientInterface
	storage  *storagev0.Storage
}

func NewSubscriber(ghClient github.GitHubClientInterface, storage *storagev0.Storage) *Subscriber {
	return &Subscriber{
		ghClient: ghClient,
		storage:  storage,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, prInfo *types.PRInfo) error {
	log.Println("Starting subscriber checker")
	var prState types.PRState
	ghState, err := s.ghClient.GetPRState(ctx, prInfo)
	if err != nil {
		return err
	}
	prState = TranslateQueryIntoState(ghState)

	s.storage.AddSubscription(prInfo, prState)
	return nil
}

func TranslateQueryIntoState(prQuery types.PRQuery) types.PRState {
	return types.PRState{
		Body:     prQuery.Repository.PullRequest.Body,
		Comments: prQuery.Repository.PullRequest.Comments.Nodes,
		Commits:  prQuery.Repository.PullRequest.Commits.Nodes,
	}
}
