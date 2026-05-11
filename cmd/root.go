package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Version: "1.0",
	Use:     "github-notifier",
	Short:   "Notifier of GitHub events on subscription",
	Long:    "This service helps you active notifications for GitHub PR events by providing the PR details",
}

func Execute() {
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(subscribeCmd)
	rootCmd.AddCommand(unsubscribeCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
