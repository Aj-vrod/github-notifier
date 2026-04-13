package poller

import (
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/pkg/slack"
	"Aj-vrod/github-notifier/pkg/subscriber"
	"Aj-vrod/github-notifier/types"
	"context"
	"log"
	"time"
)

const defaultPollInterval = 30 * time.Second

type Config struct {
	PollInterval time.Duration `envconfig:"POLL_INTERVAL" default:"30s"`
}

type Poller struct {
	Storage  *storagev0.Storage
	Config   *Config
	GHClient *github.GithubClient
	notifier *slack.SlackClient
}

func NewPoller(storage *storagev0.Storage, cfg *Config, ghClient *github.GithubClient, notifier *slack.SlackClient) *Poller {
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
		log.Printf("Checking PR: %s", prURL)

		prDetails, err := api.ParsePRURL(prURL)
		if err != nil {
			log.Printf("Error parsing PR URL %s: %v", prURL, err)
			continue
		}
		prLatestQuery, err := p.GHClient.GetPRState(ctx, prDetails)
		if err != nil {
			log.Printf("Error fetching PR info for %s: %v", prURL, err)
			continue
		}
		prLatestState := subscriber.TranslateQueryIntoState(prLatestQuery)

		if comparePRStates(prOldState, prLatestState) {
			log.Printf("Changes detected for PR %s, sending notification", prURL)
			// Update the stored state with the latest state
			p.Storage.AddSubscription(prDetails, types.PRState{
				Body:     prLatestState.Body,
				Comments: prLatestState.Comments,
				Commits:  prLatestState.Commits,
			})

			// Send a notification to Slack
			message := "Changes detected in PR: " + prURL
			if err := p.notifier.SendNotification(message); err != nil {
				log.Printf("Error sending notification for PR %s: %v", prURL, err)
			}
		} else {
			log.Printf("No changes detected for PR %s", prURL)
		}

	}

}

// Compare relevant fields to determine if there are changes
func comparePRStates(oldState, newState types.PRState) bool {
	if oldState.Body != newState.Body {
		return true
	}

	if compareComments(oldState.Comments, newState.Comments) {
		return true
	}

	if compareCommits(oldState.Commits, newState.Commits) {
		return true
	}

	return false
}

func compareComments(oldComments, newComments []types.Comment) bool {
	if len(oldComments) != len(newComments) {
		return true
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

	return false
}

func compareCommits(oldCommits, newCommits []types.CommitNode) bool {
	if len(oldCommits) != len(newCommits) {
		return true
	}

	return false
}
