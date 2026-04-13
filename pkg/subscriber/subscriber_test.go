package subscriber

import (
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/types"
	"context"
	"errors"
	"testing"
	"time"
)

// mockGitHubClient implements github.GitHubClientInterface for testing
type mockGitHubClient struct {
	GetPRStateFunc func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error)
}

func (m *mockGitHubClient) GetPRState(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
	if m.GetPRStateFunc != nil {
		return m.GetPRStateFunc(ctx, prInfo)
	}
	return types.PRQuery{}, nil
}

func TestSubscribe_Success(t *testing.T) {
	testTime := time.Now()
	storage := storagev0.NewStorage()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
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
			}, nil
		},
	}

	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	err := sub.Subscribe(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("Subscribe() error = %v, want nil", err)
	}

	// Verify the subscription was stored
	subscriptions := storage.GetAllSubscriptions()
	if len(subscriptions) != 1 {
		t.Fatalf("Expected 1 subscription, got %d", len(subscriptions))
	}

	storedState, exists := subscriptions[prInfo.URL]
	if !exists {
		t.Fatal("PR not found in storage")
	}

	if storedState.Body != "Test PR body" {
		t.Errorf("Stored body = %v, want 'Test PR body'", storedState.Body)
	}

	if len(storedState.Comments) != 1 {
		t.Errorf("Stored comments count = %d, want 1", len(storedState.Comments))
	}

	if len(storedState.Commits) != 1 {
		t.Errorf("Stored commits count = %d, want 1", len(storedState.Commits))
	}
}

func TestSubscribe_GithubError(t *testing.T) {
	storage := storagev0.NewStorage()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("GitHub API error: rate limit exceeded")
		},
	}

	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	err := sub.Subscribe(context.Background(), prInfo)
	if err == nil {
		t.Error("Subscribe() error = nil, want error when GitHub client fails")
	}

	// Verify nothing was stored
	subscriptions := storage.GetAllSubscriptions()
	if len(subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions after error, got %d", len(subscriptions))
	}
}

func TestSubscribe_PRNotFound(t *testing.T) {
	storage := storagev0.NewStorage()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("PR does not exists")
		},
	}

	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/999",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 999,
	}

	err := sub.Subscribe(context.Background(), prInfo)
	if err == nil {
		t.Error("Subscribe() error = nil, want error for non-existent PR")
	}

	if err.Error() != "PR does not exists" {
		t.Errorf("Subscribe() error = %v, want 'PR does not exists'", err)
	}
}

func TestSubscribe_EmptyPRData(t *testing.T) {
	storage := storagev0.NewStorage()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "",
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

	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	err := sub.Subscribe(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("Subscribe() error = %v, want nil for empty PR data", err)
	}

	subscriptions := storage.GetAllSubscriptions()
	storedState := subscriptions[prInfo.URL]

	if storedState.Body != "" {
		t.Errorf("Stored body = %v, want empty string", storedState.Body)
	}

	if len(storedState.Comments) != 0 {
		t.Errorf("Stored comments count = %d, want 0", len(storedState.Comments))
	}

	if len(storedState.Commits) != 0 {
		t.Errorf("Stored commits count = %d, want 0", len(storedState.Commits))
	}
}

func TestSubscribe_UpdateExistingSubscription(t *testing.T) {
	testTime := time.Now()
	storage := storagev0.NewStorage()

	// First subscription
	mockGH1 := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Original body",
						Comments: types.PRComments{
							Nodes: []types.Comment{
								{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
							},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{
								{Commit: types.Commit{Message: "Commit 1"}},
							},
						},
					},
				},
			}, nil
		},
	}

	sub1 := NewSubscriber(mockGH1, storage)
	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	err := sub1.Subscribe(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("First Subscribe() error = %v, want nil", err)
	}

	// Second subscription with updated data
	mockGH2 := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Updated body",
						Comments: types.PRComments{
							Nodes: []types.Comment{
								{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
								{Body: "Comment 2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
							},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{
								{Commit: types.Commit{Message: "Commit 1"}},
								{Commit: types.Commit{Message: "Commit 2"}},
							},
						},
					},
				},
			}, nil
		},
	}

	sub2 := NewSubscriber(mockGH2, storage)
	err = sub2.Subscribe(context.Background(), prInfo)
	if err != nil {
		t.Fatalf("Second Subscribe() error = %v, want nil", err)
	}

	// Verify the subscription was updated
	subscriptions := storage.GetAllSubscriptions()
	if len(subscriptions) != 1 {
		t.Fatalf("Expected 1 subscription, got %d", len(subscriptions))
	}

	storedState := subscriptions[prInfo.URL]
	if storedState.Body != "Updated body" {
		t.Errorf("Stored body = %v, want 'Updated body'", storedState.Body)
	}

	if len(storedState.Comments) != 2 {
		t.Errorf("Stored comments count = %d, want 2", len(storedState.Comments))
	}

	if len(storedState.Commits) != 2 {
		t.Errorf("Stored commits count = %d, want 2", len(storedState.Commits))
	}
}

func TestSubscribe_ContextCancellation(t *testing.T) {
	storage := storagev0.NewStorage()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			select {
			case <-ctx.Done():
				return types.PRQuery{}, ctx.Err()
			default:
				return types.PRQuery{}, nil
			}
		},
	}

	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := sub.Subscribe(ctx, prInfo)
	if err == nil {
		t.Error("Subscribe() error = nil, want context cancellation error")
	}
}

func TestTranslateQueryIntoState(t *testing.T) {
	testTime := time.Now()

	prQuery := types.PRQuery{
		Repository: struct {
			PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
		}{
			PullRequest: types.PRData{
				Body: "Translation test body",
				Comments: types.PRComments{
					Nodes: []types.Comment{
						{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
						{Body: "Comment 2", Author: types.Author{Login: "user2"}, CreatedAt: testTime.Add(1 * time.Hour)},
					},
				},
				Commits: types.PRCommits{
					Nodes: []types.CommitNode{
						{Commit: types.Commit{Message: "Commit message 1"}},
						{Commit: types.Commit{Message: "Commit message 2"}},
						{Commit: types.Commit{Message: "Commit message 3"}},
					},
				},
			},
		},
	}

	state := TranslateQueryIntoState(prQuery)

	if state.Body != "Translation test body" {
		t.Errorf("Translated body = %v, want 'Translation test body'", state.Body)
	}

	if len(state.Comments) != 2 {
		t.Errorf("Translated comments count = %d, want 2", len(state.Comments))
	}

	if state.Comments[0].Body != "Comment 1" {
		t.Errorf("First comment body = %v, want 'Comment 1'", state.Comments[0].Body)
	}

	if state.Comments[1].Author.Login != "user2" {
		t.Errorf("Second comment author = %v, want 'user2'", state.Comments[1].Author.Login)
	}

	if len(state.Commits) != 3 {
		t.Errorf("Translated commits count = %d, want 3", len(state.Commits))
	}

	if state.Commits[0].Commit.Message != "Commit message 1" {
		t.Errorf("First commit message = %v, want 'Commit message 1'", state.Commits[0].Commit.Message)
	}
}

func TestTranslateQueryIntoState_EmptyData(t *testing.T) {
	prQuery := types.PRQuery{
		Repository: struct {
			PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
		}{
			PullRequest: types.PRData{
				Body: "",
				Comments: types.PRComments{
					Nodes: []types.Comment{},
				},
				Commits: types.PRCommits{
					Nodes: []types.CommitNode{},
				},
			},
		},
	}

	state := TranslateQueryIntoState(prQuery)

	if state.Body != "" {
		t.Errorf("Translated body = %v, want empty string", state.Body)
	}

	if len(state.Comments) != 0 {
		t.Errorf("Translated comments count = %d, want 0", len(state.Comments))
	}

	if len(state.Commits) != 0 {
		t.Errorf("Translated commits count = %d, want 0", len(state.Commits))
	}
}
