package service

import (
	"Aj-vrod/github-notifier/config"
	"Aj-vrod/github-notifier/internal/poller"
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/pkg/slack"
	"Aj-vrod/github-notifier/pkg/subscriber"
	"context"
	"log"
	"sync"
)

type ServiceClient struct {
	Storage    *storagev0.Storage
	Subscriber *subscriber.Subscriber
}

var (
	Service     *ServiceClient
	serviceOnce sync.Once
)

func LoadService() {
	serviceOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// initiate the poller: NewPoller.Start() in a separate goroutine. Start github client there
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		log.Println("Starting GitHub Client")
		gh := github.NewClient(cfg.GithubCfg)
		log.Println("Starting Storage")
		storage := storagev0.NewStorage()
		log.Println("Starting Subscriber")
		subscriber := subscriber.NewSubscriber(gh, storage)
		log.Println("Starting Slack Client")
		notifier := slack.NewSlackClient(&cfg.SlackCfg)

		log.Println("Starting poller")
		poller := poller.NewPoller(storage, &cfg.PollerCfg, gh, notifier)
		// Handle graceful shutdown in the
		pollerShutDown := make(chan error)
		go poller.Start(ctx, pollerShutDown)

		// Wait for shutdown signal from poller
		if err := <-pollerShutDown; err != nil {
			log.Printf("Poller shutdown with error: %v", err)
		} else {
			log.Println("Poller shutdown gracefully")
		}

		Service = &ServiceClient{
			Storage:    storage,
			Subscriber: subscriber,
		}
	})
}

func (s *ServiceClient) GetServiceStorage() *storagev0.Storage {
	return s.Storage
}

func (s *ServiceClient) GetServiceSubscriber() *subscriber.Subscriber {
	return s.Subscriber
}
