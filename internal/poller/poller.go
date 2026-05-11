package poller

import (
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/pkg/slack"
	"Aj-vrod/github-notifier/pkg/subscriber"
	"Aj-vrod/github-notifier/types"
	"context"
	"fmt"
	"log"
	"time"
)

const defaultPollInterval = 30 * time.Second

type Config struct {
	PollInterval time.Duration `envconfig:"POLL_INTERVAL" default:"30s"`
	UserEmail    string        `envconfig:"USER_EMAIL" required:"true"`
	User         string        `envconfig:"USERNAME" required:"true"`
}

type Poller struct {
	Storage  *storagev0.Storage
	Config   *Config
	GHClient github.GitHubClientInterface
	notifier slack.SlackClientInterface
}

func NewPoller(storage *storagev0.Storage, cfg *Config, ghClient github.GitHubClientInterface, notifier slack.SlackClientInterface) *Poller {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}

	return &Poller{
		Storage:  storage,
		Config:   cfg,
		GHClient: ghClient,
		notifier: notifier,
	}
}

func (p *Poller) Start(ctx context.Context, shutdown chan<- error) {
	log.Println("Running poller")

	ticker := time.NewTicker(p.Config.PollInterval)
	defer ticker.Stop()
	defer close(shutdown)

	for {
		select {
		case <-ctx.Done():
			log.Println("Poller received shutdown signal")
			return
		case <-ticker.C:
			log.Println("Checking for updates")
			p.checkSubscriptions(ctx)
			log.Println("Finished checking for updates")
		}
	}

}
func (p *Poller) checkSubscriptions(ctx context.Context) {
	subscriptions := p.Storage.GetAllSubscriptions()
	for prURL, prOldState := range subscriptions {
		log.Println("Starting polling...")

		prInfo, err := api.ParsePRURL(prURL)
		if err != nil {
			log.Printf("Error parsing PR: %s/%s #%d: %v", prInfo.Owner, prInfo.Repo, prInfo.Number, err)
			continue
		}
		prLatestQuery, err := p.GHClient.GetPRState(ctx, prInfo)
		if err != nil {
			log.Printf("Error fetching PR: %s/%s #%d: %v", prInfo.Owner, prInfo.Repo, prInfo.Number, err)
			continue
		}
		prLatestState := subscriber.TranslateQueryIntoState(prLatestQuery)

		if comparePRStates(prOldState, prLatestState, p.Config) {
			log.Printf("Changes detected for PR %s/%s #%d, sending notification", prInfo.Owner, prInfo.Repo, prInfo.Number)
			// Update the stored state with the latest state
			p.Storage.Subscribe(prInfo, types.PRState{
				Body:     prLatestState.Body,
				Comments: prLatestState.Comments,
				Commits:  prLatestState.Commits,
			})

			// Send a notification to Slack
			message := fmt.Sprintf("Changes deteced in PR: %s/%s #%d", prInfo.Owner, prInfo.Repo, prInfo.Number)
			if err := p.notifier.SendNotification(message); err != nil {
				log.Printf("Error sending notification for PR: %s/%s #%d: %v", prInfo.Owner, prInfo.Repo, prInfo.Number, err)
			}
		} else {
			log.Printf("No changes detected for PR: %s/%s #%d", prInfo.Owner, prInfo.Repo, prInfo.Number)
		}

	}

}

// Compare relevant fields to determine if there are changes
func comparePRStates(oldState, newState types.PRState, cfg *Config) bool {
	if oldState.Body != newState.Body {
		return true
	}

	if oldState.ReviewDecision != newState.ReviewDecision {
		return true
	}

	if compareComments(oldState.Comments, newState.Comments, cfg.User) {
		return true
	}

	if compareCommits(oldState.Commits, newState.Commits, cfg.User) {
		return true
	}

	return false
}

func compareComments(oldComments, newComments []types.Comment, username string) bool {
	if len(oldComments) < len(newComments) {
		if newComments[len(newComments)-1].Author.Login != username {
			log.Println("List of comments changed")
			return true
		}
	}

	for i, oldComment := range oldComments {
		newComment := newComments[i]
		if oldComment.Body != newComment.Body {
			log.Printf("Comment body changed from %s to %s", oldComment.Body, newComment.Body)
			return true
		}

		if oldComment.Author.Login != newComment.Author.Login {
			log.Printf("Comment author changed from %s to %s", oldComment.Author.Login, newComment.Author.Login)
			return true
		}

		if !oldComment.CreatedAt.Equal(newComment.CreatedAt) {
			log.Printf("Comment creation time changed from %s to %s", oldComment.CreatedAt, newComment.CreatedAt)
			return true
		}
	}

	log.Println("No change in comments")
	return false
}

func compareCommits(oldCommits, newCommits []types.CommitNode, userEmail string) bool {
	if len(oldCommits) < len(newCommits) {
		if newCommits[len(newCommits)-1].Commit.Author.Email != userEmail {
			log.Println("List of commits changed")
			return true
		}
	}

	log.Println("No change in commits")
	return false
}
