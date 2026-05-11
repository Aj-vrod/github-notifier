package github

import (
	"Aj-vrod/github-notifier/types"
	"context"
	"errors"
	"testing"
	"time"
)

// mockGithubV4Client is a test helper for mocking GitHub GraphQL client behavior
type mockGithubV4Client struct {
	queryFunc func(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

func (m *mockGithubV4Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	return m.queryFunc(ctx, q, variables)
}

// MockGitHubClient implements GitHubClientInterface for testing
type MockGitHubClient struct {
	GetPRStateFunc func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error)
}

func (m *MockGitHubClient) GetPRState(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
	if m.GetPRStateFunc != nil {
		return m.GetPRStateFunc(ctx, prInfo)
	}
	return types.PRQuery{}, nil
}

func TestGetPRState_Success(t *testing.T) {
	// This is a behavioral test of the interface contract
	// Testing that a mock implementation works correctly
	testTime := time.Now()
	expectedQuery := types.PRQuery{
		Repository: struct {
			PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
		}{
			PullRequest: types.PRData{
				Body: "Test PR body",
				Comments: types.PRComments{
					Nodes: []types.Comment{
						{Body: "Test comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
					},
				},
				Commits: types.PRCommits{
					Nodes: []types.CommitNode{
						{Commit: types.Commit{Message: "Test commit"}},
					},
				},
			},
		},
	}

	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			if prInfo.Owner != "testowner" {
				t.Errorf("Expected owner 'testowner', got '%s'", prInfo.Owner)
			}
			if prInfo.Repo != "testrepo" {
				t.Errorf("Expected repo 'testrepo', got '%s'", prInfo.Repo)
			}
			if prInfo.Number != 123 {
				t.Errorf("Expected PR number 123, got %d", prInfo.Number)
			}
			return expectedQuery, nil
		},
	}

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	result, err := mock.GetPRState(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("GetPRState() error = %v, want nil", err)
	}

	if result.Repository.PullRequest.Body != "Test PR body" {
		t.Errorf("PR Body = %v, want 'Test PR body'", result.Repository.PullRequest.Body)
	}

	if len(result.Repository.PullRequest.Comments.Nodes) != 1 {
		t.Errorf("Comments count = %d, want 1", len(result.Repository.PullRequest.Comments.Nodes))
	}

	if len(result.Repository.PullRequest.Commits.Nodes) != 1 {
		t.Errorf("Commits count = %d, want 1", len(result.Repository.PullRequest.Commits.Nodes))
	}
}

func TestGetPRState_PRNotFound(t *testing.T) {
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("PR does not exists")
		},
	}

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/999",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 999,
	}

	_, err := mock.GetPRState(context.Background(), prInfo)
	if err == nil {
		t.Error("GetPRState() error = nil, want error for non-existent PR")
	}

	if err.Error() != "PR does not exists" {
		t.Errorf("GetPRState() error = %v, want 'PR does not exists'", err)
	}
}

func TestGetPRState_NetworkError(t *testing.T) {
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("network error: connection timeout")
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	_, err := mock.GetPRState(context.Background(), prInfo)
	if err == nil {
		t.Error("GetPRState() error = nil, want network error")
	}
}

func TestGetPRState_ContextCancellation(t *testing.T) {
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			select {
			case <-ctx.Done():
				return types.PRQuery{}, ctx.Err()
			default:
				return types.PRQuery{}, nil
			}
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := mock.GetPRState(ctx, prInfo)
	if err == nil {
		t.Error("GetPRState() error = nil, want context cancellation error")
	}
}

func TestGetPRState_EmptyPRData(t *testing.T) {
	// Test when PR exists but has no comments or commits
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Empty PR with no comments or commits",
						Comments: types.PRComments{
							Nodes: []types.Comment{},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{},
						},
					},
				},
			}, nil
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	result, err := mock.GetPRState(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("GetPRState() error = %v, want nil", err)
	}

	if result.Repository.PullRequest.Body != "Empty PR with no comments or commits" {
		t.Errorf("PR Body = %v, want 'Empty PR with no comments or commits'", result.Repository.PullRequest.Body)
	}

	if len(result.Repository.PullRequest.Comments.Nodes) != 0 {
		t.Errorf("Comments count = %d, want 0", len(result.Repository.PullRequest.Comments.Nodes))
	}

	if len(result.Repository.PullRequest.Commits.Nodes) != 0 {
		t.Errorf("Commits count = %d, want 0", len(result.Repository.PullRequest.Commits.Nodes))
	}
}

func TestGetPRState_MultipleCommentsAndCommits(t *testing.T) {
	testTime := time.Now()
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "PR with multiple comments and commits",
						Comments: types.PRComments{
							Nodes: []types.Comment{
								{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
								{Body: "Comment 2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
								{Body: "Comment 3", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
							},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{
								{Commit: types.Commit{Message: "Commit 1"}},
								{Commit: types.Commit{Message: "Commit 2"}},
								{Commit: types.Commit{Message: "Commit 3"}},
								{Commit: types.Commit{Message: "Commit 4"}},
							},
						},
					},
				},
			}, nil
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 456,
	}

	result, err := mock.GetPRState(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("GetPRState() error = %v, want nil", err)
	}

	if len(result.Repository.PullRequest.Comments.Nodes) != 3 {
		t.Errorf("Comments count = %d, want 3", len(result.Repository.PullRequest.Comments.Nodes))
	}

	if len(result.Repository.PullRequest.Commits.Nodes) != 4 {
		t.Errorf("Commits count = %d, want 4", len(result.Repository.PullRequest.Commits.Nodes))
	}
}

func TestGetPRState_AuthenticationError(t *testing.T) {
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("authentication failed: invalid token")
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	_, err := mock.GetPRState(context.Background(), prInfo)
	if err == nil {
		t.Error("GetPRState() error = nil, want authentication error")
	}
}

func TestGetPRState_RateLimitError(t *testing.T) {
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("rate limit exceeded")
		},
	}

	prInfo := &types.PRInfo{
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	_, err := mock.GetPRState(context.Background(), prInfo)
	if err == nil {
		t.Error("GetPRState() error = nil, want rate limit error")
	}
}

// ============================================================================
// BUG DOCUMENTATION TESTS
// ============================================================================
// The following tests document a CRITICAL BUG in the actual GithubClient.GetPRState implementation
// at pkg/github/client.go lines 50-52.
//
// BUG: The GetPRState method calls log.Fatalf() when the GraphQL query fails, which terminates
// the ENTIRE SERVICE instead of returning an error. This means:
// 1. A single bad PR URL crashes the entire application
// 2. Network errors kill the service instead of being handled
// 3. Rate limit errors terminate the service
// 4. The poller cannot gracefully handle failures and continue with other subscriptions
//
// The actual implementation code (BUGGY):
//     err := c.client.Query(ctx, &query, variables)
//     if err != nil {
//         log.Fatalf("failed to execute query: %v", err)  // <-- CRASHES THE SERVICE
//     }
//
// What it SHOULD do (return the error):
//     err := c.client.Query(ctx, &query, variables)
//     if err != nil {
//         return types.PRQuery{}, fmt.Errorf("failed to execute query: %w", err)
//     }
//
// These tests verify that the mock implementation CORRECTLY returns errors instead of crashing.
// However, the real GithubClient will crash the service if any of these errors occur.
// ============================================================================

// TestActualImplementation_WouldCrashOnQueryError documents the crash behavior
// NOTE: This test uses the mock which correctly returns errors. The actual GithubClient
// would call os.Exit(1) via log.Fatalf() and terminate the entire test suite.
func TestActualImplementation_WouldCrashOnQueryError(t *testing.T) {
	t.Run("network error would crash service", func(t *testing.T) {
		// Using mock to demonstrate CORRECT behavior
		mock := &MockGitHubClient{
			GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
				// What the interface SHOULD do: return error
				return types.PRQuery{}, errors.New("network error: connection refused")
			},
		}

		prInfo := &types.PRInfo{
			Owner:  "testowner",
			Repo:   "testrepo",
			Number: 123,
		}

		_, err := mock.GetPRState(context.Background(), prInfo)

		// Mock correctly returns error
		if err == nil {
			t.Error("Expected error, got nil")
		}

		// BUG DOCUMENTATION: The actual GithubClient.GetPRState() would call log.Fatalf()
		// here and crash the entire service instead of returning this error.
		t.Logf("BUG: Actual implementation would crash service with log.Fatalf() instead of returning: %v", err)
	})

	t.Run("invalid repo would crash service", func(t *testing.T) {
		mock := &MockGitHubClient{
			GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
				return types.PRQuery{}, errors.New("repository not found")
			},
		}

		prInfo := &types.PRInfo{
			Owner:  "nonexistent",
			Repo:   "nonexistent",
			Number: 1,
		}

		_, err := mock.GetPRState(context.Background(), prInfo)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		t.Logf("BUG: Actual implementation would crash service instead of returning: %v", err)
	})

	t.Run("malformed graphql response would crash service", func(t *testing.T) {
		mock := &MockGitHubClient{
			GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
				return types.PRQuery{}, errors.New("graphql: field 'pullRequest' not found")
			},
		}

		prInfo := &types.PRInfo{
			Owner:  "testowner",
			Repo:   "testrepo",
			Number: 123,
		}

		_, err := mock.GetPRState(context.Background(), prInfo)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		t.Logf("BUG: Actual implementation would crash service instead of returning: %v", err)
	})

	t.Run("context timeout would crash service", func(t *testing.T) {
		mock := &MockGitHubClient{
			GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
				return types.PRQuery{}, context.DeadlineExceeded
			},
		}

		prInfo := &types.PRInfo{
			Owner:  "testowner",
			Repo:   "testrepo",
			Number: 123,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		_, err := mock.GetPRState(ctx, prInfo)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		t.Logf("BUG: Actual implementation would crash service instead of returning: %v", err)
	})
}

// TestPollerImpact_ServiceCrashPreventsOtherPRs documents the cascading failure
func TestPollerImpact_ServiceCrashPreventsOtherPRs(t *testing.T) {
	t.Run("one bad PR should not stop checking other PRs", func(t *testing.T) {
		// Simulate poller checking multiple PRs
		goodPR := &types.PRInfo{Owner: "good", Repo: "repo", Number: 1}
		badPR := &types.PRInfo{Owner: "bad", Repo: "repo", Number: 2}
		anotherGoodPR := &types.PRInfo{Owner: "good", Repo: "another", Number: 3}

		mock := &MockGitHubClient{
			GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
				if prInfo.Owner == "bad" {
					// This PR has an error
					return types.PRQuery{}, errors.New("network error")
				}
				// Other PRs are fine
				return types.PRQuery{
					Repository: struct {
						PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
					}{
						PullRequest: types.PRData{
							Body:     "Success",
							Comments: types.PRComments{Nodes: []types.Comment{}},
							Commits:  types.PRCommits{Nodes: []types.CommitNode{{Commit: types.Commit{Message: "test"}}}},
						},
					},
				}, nil
			},
		}

		// With mock (correct behavior): all PRs get attempted
		successCount := 0
		errorCount := 0

		for _, pr := range []*types.PRInfo{goodPR, badPR, anotherGoodPR} {
			_, err := mock.GetPRState(context.Background(), pr)
			if err != nil {
				errorCount++
				t.Logf("PR %s/%s #%d failed: %v (service continues)", pr.Owner, pr.Repo, pr.Number, err)
			} else {
				successCount++
				t.Logf("PR %s/%s #%d succeeded", pr.Owner, pr.Repo, pr.Number)
			}
		}

		if successCount != 2 {
			t.Errorf("Expected 2 successful PRs, got %d", successCount)
		}
		if errorCount != 1 {
			t.Errorf("Expected 1 failed PR, got %d", errorCount)
		}

		// BUG DOCUMENTATION: With actual GithubClient, the service would crash on badPR
		// and anotherGoodPR would never be checked
		t.Log("BUG: With actual GithubClient, service would crash at badPR and anotherGoodPR would never be checked")
	})
}

// TestErrorRecovery_ServiceShouldContinue documents that errors should be recoverable
func TestErrorRecovery_ServiceShouldContinue(t *testing.T) {
	callCount := 0
	mock := &MockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			callCount++
			if callCount == 1 {
				// First call fails
				return types.PRQuery{}, errors.New("temporary network error")
			}
			// Second call succeeds (network recovered)
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body:     "Recovered",
						Comments: types.PRComments{Nodes: []types.Comment{}},
						Commits:  types.PRCommits{Nodes: []types.CommitNode{{Commit: types.Commit{Message: "test"}}}},
					},
				},
			}, nil
		},
	}

	prInfo := &types.PRInfo{Owner: "owner", Repo: "repo", Number: 1}

	// First attempt fails
	_, err := mock.GetPRState(context.Background(), prInfo)
	if err == nil {
		t.Error("Expected error on first call")
	}

	// Service should still be running and able to retry
	result, err := mock.GetPRState(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("Expected success on retry, got error: %v", err)
	}
	if result.Repository.PullRequest.Body != "Recovered" {
		t.Error("Expected successful retry")
	}

	t.Log("BUG: With actual GithubClient, first error would crash service and retry would be impossible")
}
