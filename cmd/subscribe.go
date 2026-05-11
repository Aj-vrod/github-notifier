package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"Aj-vrod/github-notifier/types"

	"github.com/spf13/cobra"
)

var validArgs = []string{
	"org",
	"repo",
	"pr",
}

var subscribeCmd = &cobra.Command{
	Use:       "sub",
	Short:     "Subscribe a PR",
	Long:      "Subscribe PR changes to receive notifications",
	Args:      cobra.MinimumNArgs(3),
	ValidArgs: validArgs,
	Example:   "... sub org=company repo=my-repo pr=123",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateArgs(args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var reqArgs *types.SubArgs
		parseArgs(args, &reqArgs)
		return subscribe(ctx, reqArgs)
	},
}

func validateArgs(args []string) error {
	for i, v := range validArgs {
		if slices.IndexFunc(args, func(s string) bool { return v == strings.Split(s, "=")[0] }) == -1 {
			return fmt.Errorf("Missing %s argument", validArgs[i])
		}
	}

	return nil
}

func parseArgs(args []string, reqArgs **types.SubArgs) {
	var org, repo, pr string
	for _, v := range args {
		pair := strings.Split(v, "=")
		switch pair[0] {
		case "org":
			org = pair[1]
		case "repo":
			repo = pair[1]
		case "pr":
			pr = pair[1]
		}
	}
	*reqArgs = &types.SubArgs{
		Org:  org,
		Repo: repo,
		PR:   pr,
	}
}

func subscribe(ctx context.Context, args *types.SubArgs) error {
	payload := fmt.Sprintf(`{"pr_url": "https://github.com/%s/%s/pull/%s"}`, args.Org, args.Repo, args.PR)
	r := bytes.NewReader([]byte(payload))
	requestURL := fmt.Sprintf("http://localhost:%d/api/v1/subscribe", ServerPort)
	req, err := http.NewRequest(http.MethodPost, requestURL, r)
	if err != nil {
		fmt.Printf("Could not create request: %s\n", err)
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not execute the request: %s\n", err)
		return err
	}
	if res.StatusCode != http.StatusCreated {
		fmt.Printf("Record was not created: %s\n", err)
		return err
	}
	fmt.Println("Record created successfully")
	return nil
}
