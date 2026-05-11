package cmd

import (
	"Aj-vrod/github-notifier/pkg/api"
	"Aj-vrod/github-notifier/pkg/service"
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
		service.LoadService()
		httpRun()
	},
}

func httpRun() {
	subscriber := service.Service.GetServiceSubscriber()
	storage := service.Service.GetServiceStorage()

	// Create the API server
	server := api.NewServer(ServerPort, subscriber, storage)

	// Start the API server in a separate goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

}
