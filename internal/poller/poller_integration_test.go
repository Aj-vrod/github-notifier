package poller

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
	callCount      int
}

func (m *mockGitHubClient) GetPRState(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
	m.callCount++
	if m.GetPRStateFunc != nil {
		return m.GetPRStateFunc(ctx, prInfo)
	}
	return types.PRQuery{}, nil
}

// mockSlackClient implements slack.SlackClientInterface for testing
type mockSlackClient struct {
	SendNotificationFunc func(message string) error
	callCount            int
	lastMessage          string
}

func (m *mockSlackClient) SendNotification(message string) error {
	m.callCount++
	m.lastMessage = message
	if m.SendNotificationFunc != nil {
		return m.SendNotificationFunc(message)
	}
	return nil
}

func TestCheckSubscriptions_DetectsBodyChange(t *testing.T) {
	storage := storagev0.NewStorage()

	// Add initial subscription
	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body:     "Old body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "New body", // Changed
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

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	p.checkSubscriptions(context.Background())

	if mockSlack.callCount != 1 {
		t.Errorf("Slack notification count = %d, want 1", mockSlack.callCount)
	}

	if mockSlack.lastMessage != "Changes detected in PR: "+prInfo.URL {
		t.Errorf("Notification message = %v, want 'Changes detected in PR: %s'", mockSlack.lastMessage, prInfo.URL)
	}

	// Verify storage was updated
	subscriptions := storage.GetAllSubscriptions()
	if subscriptions[prInfo.URL].Body != "New body" {
		t.Errorf("Updated body = %v, want 'New body'", subscriptions[prInfo.URL].Body)
	}
}

func TestCheckSubscriptions_DetectsCommentChange(t *testing.T) {
	testTime := time.Now()
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body: "Same body",
		Comments: []types.Comment{
			{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
		},
		Commits: []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Same body",
						Comments: types.PRComments{
							Nodes: []types.Comment{
								{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
								{Body: "Comment 2", Author: types.Author{Login: "user2"}, CreatedAt: testTime}, // New comment
							},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{},
						},
					},
				},
			}, nil
		},
	}

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	p.checkSubscriptions(context.Background())

	if mockSlack.callCount != 1 {
		t.Errorf("Slack notification count = %d, want 1", mockSlack.callCount)
	}
}

func TestCheckSubscriptions_DetectsCommitChange(t *testing.T) {
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body:     "Same body",
		Comments: []types.Comment{},
		Commits: []types.CommitNode{
			{Commit: types.Commit{Message: "Commit 1"}},
		},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Same body",
						Comments: types.PRComments{
							Nodes: []types.Comment{},
						},
						Commits: types.PRCommits{
							Nodes: []types.CommitNode{
								{Commit: types.Commit{Message: "Commit 1"}},
								{Commit: types.Commit{Message: "Commit 2"}}, // New commit
							},
						},
					},
				},
			}, nil
		},
	}

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	p.checkSubscriptions(context.Background())

	if mockSlack.callCount != 1 {
		t.Errorf("Slack notification count = %d, want 1", mockSlack.callCount)
	}
}

func TestCheckSubscriptions_NoChanges(t *testing.T) {
	testTime := time.Now()
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body: "Same body",
		Comments: []types.Comment{
			{Body: "Comment 1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
		},
		Commits: []types.CommitNode{
			{Commit: types.Commit{Message: "Commit 1"}},
		},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Same body",
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

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	p.checkSubscriptions(context.Background())

	if mockSlack.callCount != 0 {
		t.Errorf("Slack notification count = %d, want 0 (no changes)", mockSlack.callCount)
	}
}

func TestCheckSubscriptions_MultipleSubscriptions(t *testing.T) {
	storage := storagev0.NewStorage()

	// Add multiple subscriptions
	prInfo1 := &types.PRInfo{
		URL:    "https://github.com/owner1/repo1/pull/1",
		Owner:  "owner1",
		Repo:   "repo1",
		Number: 1,
	}
	prInfo2 := &types.PRInfo{
		URL:    "https://github.com/owner2/repo2/pull/2",
		Owner:  "owner2",
		Repo:   "repo2",
		Number: 2,
	}

	storage.AddSubscription(prInfo1, types.PRState{
		Body:     "Body 1",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})
	storage.AddSubscription(prInfo2, types.PRState{
		Body:     "Body 2",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			// Both PRs have changes
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Updated body for " + prInfo.Owner, // Changed
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

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	p.checkSubscriptions(context.Background())

	if mockGH.callCount != 2 {
		t.Errorf("GitHub API call count = %d, want 2", mockGH.callCount)
	}

	if mockSlack.callCount != 2 {
		t.Errorf("Slack notification count = %d, want 2", mockSlack.callCount)
	}
}

func TestCheckSubscriptions_GithubError(t *testing.T) {
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body:     "Body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{}, errors.New("GitHub API error")
		},
	}

	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	// Should not panic, just log error
	p.checkSubscriptions(context.Background())

	// No notification should be sent when GitHub fails
	if mockSlack.callCount != 0 {
		t.Errorf("Slack notification count = %d, want 0 (GitHub error)", mockSlack.callCount)
	}
}

func TestCheckSubscriptions_SlackError(t *testing.T) {
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body:     "Old body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "New body", // Changed
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

	mockSlack := &mockSlackClient{
		SendNotificationFunc: func(message string) error {
			return errors.New("Slack webhook error")
		},
	}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	// Should not panic, just log error
	p.checkSubscriptions(context.Background())

	// Notification was attempted
	if mockSlack.callCount != 1 {
		t.Errorf("Slack notification attempts = %d, want 1", mockSlack.callCount)
	}

	// Storage should still be updated even if Slack fails
	subscriptions := storage.GetAllSubscriptions()
	if subscriptions[prInfo.URL].Body != "New body" {
		t.Errorf("Storage not updated after Slack error, body = %v, want 'New body'", subscriptions[prInfo.URL].Body)
	}
}

func TestCheckSubscriptions_InvalidPRURL(t *testing.T) {
	storage := storagev0.NewStorage()

	// Manually add subscription with invalid URL format (bypassing normal validation)
	storage.AddSubscription(&types.PRInfo{URL: "invalid-url"}, types.PRState{
		Body:     "Body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{}
	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 1 * time.Second}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	// Should not panic, just log error
	p.checkSubscriptions(context.Background())

	// No GitHub call should be made for invalid URL
	if mockGH.callCount != 0 {
		t.Errorf("GitHub API call count = %d, want 0 (invalid URL)", mockGH.callCount)
	}
}

func TestStart_ContextCancellation(t *testing.T) {
	storage := storagev0.NewStorage()
	mockGH := &mockGitHubClient{}
	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 100 * time.Millisecond}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan error)

	// Start poller in goroutine
	go p.Start(ctx, shutdown)

	// Cancel after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for shutdown
	select {
	case <-shutdown:
		// Shutdown completed successfully
	case <-time.After(1 * time.Second):
		t.Error("Poller did not shut down within timeout")
	}
}

func TestStart_TickerFires(t *testing.T) {
	storage := storagev0.NewStorage()

	prInfo := &types.PRInfo{
		URL:    "https://github.com/testowner/testrepo/pull/123",
		Owner:  "testowner",
		Repo:   "testrepo",
		Number: 123,
	}
	storage.AddSubscription(prInfo, types.PRState{
		Body:     "Body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	})

	mockGH := &mockGitHubClient{
		GetPRStateFunc: func(ctx context.Context, prInfo *types.PRInfo) (types.PRQuery, error) {
			return types.PRQuery{
				Repository: struct {
					PullRequest types.PRData `graphql:"pullRequest(number: $prNumber)"`
				}{
					PullRequest: types.PRData{
						Body: "Body",
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
	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 100 * time.Millisecond}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan error)

	go p.Start(ctx, shutdown)

	// Wait for at least 2 ticks
	time.Sleep(250 * time.Millisecond)
	cancel()

	<-shutdown

	// Verify multiple GitHub API calls were made (ticker fired multiple times)
	if mockGH.callCount < 2 {
		t.Errorf("GitHub API call count = %d, want at least 2 (ticker should fire multiple times)", mockGH.callCount)
	}
}

func TestNewPoller_DefaultPollInterval(t *testing.T) {
	storage := storagev0.NewStorage()
	mockGH := &mockGitHubClient{}
	mockSlack := &mockSlackClient{}

	cfg := &Config{PollInterval: 0} // No interval set
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	expectedDefault := 30 * time.Second
	if p.Config.PollInterval != expectedDefault {
		t.Errorf("PollInterval = %v, want default %v", p.Config.PollInterval, expectedDefault)
	}
}

func TestNewPoller_CustomPollInterval(t *testing.T) {
	storage := storagev0.NewStorage()
	mockGH := &mockGitHubClient{}
	mockSlack := &mockSlackClient{}

	customInterval := 45 * time.Second
	cfg := &Config{PollInterval: customInterval}
	p := NewPoller(storage, cfg, mockGH, mockSlack)

	if p.Config.PollInterval != customInterval {
		t.Errorf("PollInterval = %v, want %v", p.Config.PollInterval, customInterval)
	}
}
