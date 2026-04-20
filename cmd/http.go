package cmd

import (
	"Aj-vrod/github-notifier/config"
	"Aj-vrod/github-notifier/internal/poller"
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/pkg/slack"
	"Aj-vrod/github-notifier/pkg/subscriber"
	"context"
	"log"

	"github.com/spf13/cobra"
)

const (
	// ServerPort is the port number the API server will listen on
	ServerPort = 8001
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts the server",
	Run: func(cmd *cobra.Command, args []string) {
		httpRun()
	},
}

func httpRun() {
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

	// Create the API server
	server := api.NewServer(ServerPort, subscriber, storage)

	// Start the API server in a separate goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal from poller
	if err := <-pollerShutDown; err != nil {
		log.Printf("Poller shutdown with error: %v", err)
	} else {
		log.Println("Poller shutdown gracefully")
	}
}
