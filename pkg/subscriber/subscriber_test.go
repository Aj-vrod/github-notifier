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

// ============================================================================
// EDGE CASE TESTS
// ============================================================================
// The following tests cover edge cases that may not be handled properly

// TestSubscribe_PRWith100Comments tests the pagination boundary
func TestSubscribe_PRWith100Comments(t *testing.T) {
	storage := storagev0.NewStorage()
	testTime := time.Now()

	// GitHub GraphQL queries limit comments to first 100 (hardcoded in query)
	comments := make([]types.Comment, 100)
	for i := 0; i < 100; i++ {
		comments[i] = types.Comment{
			Body:      "Comment " + string(rune(i)),
			Author:    types.Author{Login: "user" + string(rune(i))},
			CreatedAt: testTime,
		}
	}

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "PR with 100 comments",
						Comments: types.PRComments{
							Nodes: comments,
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
		t.Fatalf("Subscribe() error = %v, want nil", err)
	}

	subscriptions := storage.GetAllSubscriptions()
	storedState := subscriptions[prInfo.URL]

	if len(storedState.Comments) != 100 {
		t.Errorf("Stored comments count = %d, want 100", len(storedState.Comments))
	}

	// NOTE: If PR actually has > 100 comments, only first 100 are retrieved
	// This is a limitation of the hardcoded GraphQL query
	t.Log("LIMITATION: GraphQL query hardcoded to first 100 comments")
	t.Log("If PR has > 100 comments, additional comments are not tracked")
}

// TestSubscribe_PRWith100Commits tests the pagination boundary for commits
func TestSubscribe_PRWith100Commits(t *testing.T) {
	storage := storagev0.NewStorage()

	// GitHub GraphQL queries limit commits to first 100 (hardcoded in query)
	commits := make([]types.CommitNode, 100)
	for i := 0; i < 100; i++ {
		commits[i] = types.CommitNode{
			Commit: types.Commit{
				Message: "Commit " + string(rune(i)),
			},
		}
	}

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "PR with 100 commits",
						Comments: types.PRComments{
							Nodes: []types.Comment{},
						},
						Commits: types.PRCommits{
							Nodes: commits,
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

	subscriptions := storage.GetAllSubscriptions()
	storedState := subscriptions[prInfo.URL]

	if len(storedState.Commits) != 100 {
		t.Errorf("Stored commits count = %d, want 100", len(storedState.Commits))
	}

	t.Log("LIMITATION: GraphQL query hardcoded to first 100 commits")
	t.Log("If PR has > 100 commits, additional commits are not tracked")
}

// TestSubscribe_NilAuthorFields tests edge case with nil authors
func TestSubscribe_NilAuthorFields(t *testing.T) {
	storage := storagev0.NewStorage()
	testTime := time.Now()

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "PR with empty author",
						Comments: types.PRComments{
							Nodes: []types.Comment{
								{Body: "Comment with empty author", Author: types.Author{Login: ""}, CreatedAt: testTime},
							},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{
								{Commit: types.Commit{Message: "Commit with empty author", Author: types.GitAuthor{Email: ""}}},
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

	subscriptions := storage.GetAllSubscriptions()
	storedState := subscriptions[prInfo.URL]

	// Verify empty authors are stored
	if storedState.Comments[0].Author.Login != "" {
		t.Errorf("Expected empty author login, got '%s'", storedState.Comments[0].Author.Login)
	}

	if storedState.Commits[0].Commit.Author.Email != "" {
		t.Errorf("Expected empty author email, got '%s'", storedState.Commits[0].Commit.Author.Email)
	}

	t.Log("RISK: Empty author fields may cause issues in poller's username filtering")
}

// TestSubscribe_ReviewDecision tests that review decisions are stored
func TestSubscribe_ReviewDecision(t *testing.T) {
	storage := storagev0.NewStorage()

	testCases := []struct {
		name           string
		reviewDecision string
	}{
		{"approved", "APPROVED"},
		{"changes requested", "CHANGES_REQUESTED"},
		{"review required", "REVIEW_REQUIRED"},
		{"empty decision", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGH := &mockGitHubClient{
				GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
					return types.PRQuery{
						Repository: struct {
							PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
						}{
							PullRequest: types.PRData{
								Body:           "PR body",
								Comments:       types.PRComments{Nodes: []types.Comment{}},
								Commits:        types.PRCommits{Nodes: []types.CommitNode{}},
								ReviewDecision: tc.reviewDecision,
							},
						},
					}, nil
				},
			}

			sub := NewSubscriber(mockGH, storage)
			prInfo := &types.PRInfo{
				URL:    "https://github.com/testowner/testrepo/pull/" + tc.name,
				Owner:  "testowner",
				Repo:   "testrepo",
				Number: 123,
			}

			err := sub.Subscribe(context.Background(), prInfo)
			if err != nil {
				t.Fatalf("Subscribe() error = %v, want nil", err)
			}

			subscriptions := storage.GetAllSubscriptions()
			storedState := subscriptions[prInfo.URL]

			if storedState.ReviewDecision != tc.reviewDecision {
				t.Errorf("ReviewDecision = %v, want %v", storedState.ReviewDecision, tc.reviewDecision)
			}
		})
	}
}

// TestUnsubscribe tests the unsubscribe functionality
func TestUnsubscribe(t *testing.T) {
	storage := storagev0.NewStorage()
	mockGH := &mockGitHubClient{}
	sub := NewSubscriber(mockGH, storage)

	// Add a subscription first
	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	prState := types.PRState{
		Body:     "Test",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	}
	storage.Subscribe(prInfo, prState)

	// Verify subscription exists
	subscriptions := storage.GetAllSubscriptions()
	if len(subscriptions) != 1 {
		t.Fatalf("Expected 1 subscription before unsubscribe, got %d", len(subscriptions))
	}

	// Unsubscribe
	sub.Unsubscribe(context.Background(), prInfo)

	// Verify subscription was removed
	subscriptions = storage.GetAllSubscriptions()
	if len(subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions after unsubscribe, got %d", len(subscriptions))
	}
}

// TestUnsubscribe_NonExistentPR tests unsubscribing a PR that isn't subscribed
func TestUnsubscribe_NonExistentPR(t *testing.T) {
	storage := storagev0.NewStorage()
	mockGH := &mockGitHubClient{}
	sub := NewSubscriber(mockGH, storage)

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/999",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 999,
	}

	// Unsubscribe non-existent PR (should not crash)
	sub.Unsubscribe(context.Background(), prInfo)

	// Verify storage is still empty
	subscriptions := storage.GetAllSubscriptions()
	if len(subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", len(subscriptions))
	}

	t.Log("OK: Unsubscribing non-existent PR does not crash")
}

// TestTranslateQueryIntoState_NilNodes tests translation with nil nodes
func TestTranslateQueryIntoState_NilNodes(t *testing.T) {
	t.Run("nil comments nodes", func(t *testing.T) {
		prQuery := types.PRQuery{
			Repository: struct {
				PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
			}{
				PullRequest: types.PRData{
					Body: "Test",
					Comments: types.PRComments{
						Nodes: nil, // nil instead of empty slice
					},
					Commits: types.PRCommits{
						Nodes: []types.CommitNode{},
					},
				},
			},
		}

		state := TranslateQueryIntoState(prQuery)

		if state.Comments == nil {
			t.Log("Comments is nil (may cause issues in comparisons)")
		}
		if len(state.Comments) != 0 {
			t.Errorf("Expected 0 comments, got %d", len(state.Comments))
		}
	})

	t.Run("nil commits nodes", func(t *testing.T) {
		prQuery := types.PRQuery{
			Repository: struct {
				PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
			}{
				PullRequest: types.PRData{
					Body: "Test",
					Comments: types.PRComments{
						Nodes: []types.Comment{},
					},
					Commits: types.PRCommits{
						Nodes: nil, // nil instead of empty slice
					},
				},
			},
		}

		state := TranslateQueryIntoState(prQuery)

		// NOTE: This is how GitHub API indicates PR doesn't exist
		// See pkg/github/client.go line 55
		if state.Commits == nil {
			t.Log("Commits is nil - indicates PR doesn't exist per client.go check")
		}
	})
}
