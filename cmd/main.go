package main

import (
	"log"

	"Aj-vrod/github-notifier/pkg/api"
)

const (
	// ServerPort is the port number the API server will listen on
	ServerPort = 8001
)

func main() {
	server := api.NewServer(ServerPort)

	// Start the server
	log.Fatal(server.Start())
}
