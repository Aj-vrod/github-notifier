package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "github-notifier",
	Short: "Notifier of Github events on subscription",
	Long:  "This service helps you active notifications for PR events by providing the PR URL",
}

func Execute() {
	rootCmd.AddCommand(httpCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
