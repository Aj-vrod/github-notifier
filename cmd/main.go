package main

import (
	"log"

	"Aj-vrod/github-notifier/config"
	"Aj-vrod/github-notifier/internal/storagev0"

	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/github"
	"Aj-vrod/github-notifier/pkg/subscriber"
)

const (
	// ServerPort is the port number the API server will listen on
	ServerPort = 8001
)

func main() {
	// initiate the poller: NewPoller.Start() in a separate goroutine. Start github client there
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	gh := github.NewClient(cfg.GithubCfg)
	subscriber := subscriber.NewSubscriber(gh)
	storage := storagev0.NewStorage()

	server := api.NewServer(ServerPort, subscriber, storage)

	// Start the server
	log.Fatal(server.Start())
}
