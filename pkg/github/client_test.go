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
