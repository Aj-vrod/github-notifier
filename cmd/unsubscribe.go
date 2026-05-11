package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"

	"Aj-vrod/github-notifier/types"

	"github.com/spf13/cobra"
)

var unsubscribeCmd = &cobra.Command{
	Use:       "unsub",
	Short:     "Unsubscribe a PR",
	Long:      "Unsubscribe PR changes to stop receiving notifications",
	Args:      cobra.MinimumNArgs(3),
	ValidArgs: validArgs,
	Example:   "... unsub org=company repo=my-repo pr=123",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateArgs(args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var reqArgs *types.SubArgs
		parseArgs(args, &reqArgs)
		return unsubscribe(ctx, reqArgs)
	},
}

func unsubscribe(ctx context.Context, args *types.SubArgs) error {
	payload := fmt.Sprintf(`{"pr_url": "https://github.com/%s/%s/pull/%s"}`, args.Org, args.Repo, args.PR)
	r := bytes.NewReader([]byte(payload))
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/subscribe", ServerPort)
	req, err := http.NewRequest(http.MethodDelete, requestURL, r)
	if err != nil {
		return fmt.Errorf("could not create request: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute the request: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("record was not removed: %s", err)
	}
	log.Println("Record removed successfully")
	return nil
}
