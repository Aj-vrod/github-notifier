package api

import (
	"Aj-vrod/github-notifier/types"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Validation error messages as constants for consistency
const (
	errProtocolNotHTTPS      = "pr_url must use https protocol"
	errDomainNotGitHub       = "pr_url must be from github.com domain"
	errInvalidPathFormat     = "pr_url must match format: https://github.com/{owner}/{repo}/pull/{number}"
	errOwnerInvalidFormat    = "owner must contain only alphanumeric characters and hyphens"
	errOwnerTooLong          = "owner must not exceed 39 characters"
	errRepoInvalidFormat     = "repo must contain only alphanumeric characters, hyphens, underscores, and dots"
	errRepoTooLong           = "repo must not exceed 100 characters"
	errPRNumberInvalidFormat = "pull request number must be a positive integer"
)

// Regular expressions for validation
var (
	// Owner: alphanumeric and hyphens only
	ownerRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	// Repo: alphanumeric, hyphens, underscores, and dots
	repoRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// ParsePRURL parses a GitHub PR URL and extracts the owner, repo, and PR number
// Returns an error if the URL format is invalid
func ParsePRURL(prURL string) (*types.PRInfo, error) {
	// Parse the URL
	u, err := url.Parse(prURL)
	if err != nil {
		return nil, fmt.Errorf(errInvalidPathFormat)
	}

	// Validate protocol
	if u.Scheme != "https" {
		return nil, fmt.Errorf(errProtocolNotHTTPS)
	}

	// Validate domain
	if u.Host != "github.com" {
		return nil, fmt.Errorf(errDomainNotGitHub)
	}

	// Parse path: /{owner}/{repo}/pull/{number}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	// Must have exactly 4 parts: owner, repo, "pull", number
	if len(parts) != 4 || parts[2] != "pull" {
		return nil, fmt.Errorf(errInvalidPathFormat)
	}

	owner := parts[0]
	repo := parts[1]
	numberStr := parts[3]

	// Validate owner
	if err := validateOwner(owner); err != nil {
		return nil, err
	}

	// Validate repo
	if err := validateRepo(repo); err != nil {
		return nil, err
	}

	// Validate and parse PR number
	number, err := validatePRNumber(numberStr)
	if err != nil {
		return nil, err
	}

	return &types.PRInfo{
		URL:    prURL,
		Owner:  owner,
		Repo:   repo,
		Number: number,
	}, nil
}

// ValidatePRURL validates a GitHub PR URL format and returns an error if invalid
// This is a convenience function that wraps ParsePRURL
func ValidatePRURL(prURL string) error {
	_, err := ParsePRURL(prURL)
	return err
}

// validateOwner checks if the owner name is valid according to GitHub rules
func validateOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf(errInvalidPathFormat)
	}

	if !ownerRegex.MatchString(owner) {
		return fmt.Errorf(errOwnerInvalidFormat)
	}

	if len(owner) > 39 {
		return fmt.Errorf(errOwnerTooLong)
	}

	return nil
}

// validateRepo checks if the repo name is valid according to GitHub rules
func validateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf(errInvalidPathFormat)
	}

	if !repoRegex.MatchString(repo) {
		return fmt.Errorf(errRepoInvalidFormat)
	}

	if len(repo) > 100 {
		return fmt.Errorf(errRepoTooLong)
	}

	return nil
}

// validatePRNumber checks if the PR number is valid (positive integer)
func validatePRNumber(numberStr string) (int, error) {
	if numberStr == "" {
		return 0, fmt.Errorf(errInvalidPathFormat)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf(errPRNumberInvalidFormat)
	}

	return number, nil
}
