package main

import (
	"context"
	"fmt"
	"log"

	"Aj-vrod/github-notifier/config"
	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/github"

	"github.com/shurcooL/githubv4"
)

const (
	// ServerPort is the port number the API server will listen on
	ServerPort = 8001
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	gh := github.NewClient(cfg.GithubCfg)

	// Example query to verify GitHub client is working
	var query struct {
		Viewer struct {
			Login     githubv4.String
			CreatedAt githubv4.DateTime
		}
	}

	err = gh.Query(context.Background(), &query, nil)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}
	fmt.Println("    Login:", query.Viewer.Login)
	fmt.Println("CreatedAt:", query.Viewer.CreatedAt)

	server := api.NewServer(ServerPort)

	// Start the server
	log.Fatal(server.Start())
}
