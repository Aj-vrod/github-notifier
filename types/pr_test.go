package types

import (
	"testing"
	"time"
)

// TestPRInfo tests the PRInfo struct and its fields
func TestPRInfo(t *testing.T) {
	t.Run("create PRInfo with all fields", func(t *testing.T) {
		prInfo := PRInfo{
			URL:    "https://github.com/owner/repo/pull/123",
			Owner:  "owner",
			Repo:   "repo",
			Number: 123,
		}

		if prInfo.URL != "https://github.com/owner/repo/pull/123" {
			t.Errorf("URL = %v, want 'https://github.com/owner/repo/pull/123'", prInfo.URL)
		}
		if prInfo.Owner != "owner" {
			t.Errorf("Owner = %v, want 'owner'", prInfo.Owner)
		}
		if prInfo.Repo != "repo" {
			t.Errorf("Repo = %v, want 'repo'", prInfo.Repo)
		}
		if prInfo.Number != 123 {
			t.Errorf("Number = %v, want 123", prInfo.Number)
		}
	})

	t.Run("PRInfo with empty fields", func(t *testing.T) {
		prInfo := PRInfo{
			URL:    "",
			Owner:  "",
			Repo:   "",
			Number: 0,
		}

		if prInfo.URL != "" {
			t.Errorf("Expected empty URL, got %v", prInfo.URL)
		}
		if prInfo.Number != 0 {
			t.Errorf("Expected 0 for Number, got %v", prInfo.Number)
		}
	})

	t.Run("PRInfo with special characters", func(t *testing.T) {
		prInfo := PRInfo{
			URL:    "https://github.com/org-with-dash/repo_with_underscore/pull/999",
			Owner:  "org-with-dash",
			Repo:   "repo_with_underscore",
			Number: 999,
		}

		if prInfo.Owner != "org-with-dash" {
			t.Errorf("Owner = %v, want 'org-with-dash'", prInfo.Owner)
		}
		if prInfo.Repo != "repo_with_underscore" {
			t.Errorf("Repo = %v, want 'repo_with_underscore'", prInfo.Repo)
		}
	})
}

// TestPRState tests the PRState struct
func TestPRState(t *testing.T) {
	t.Run("create PRState with all fields", func(t *testing.T) {
		testTime := time.Now()
		prState := PRState{
			Body: "PR description",
			Comments: []Comment{
				{Body: "Comment 1", Author: Author{Login: "user1"}, CreatedAt: testTime},
			},
			Commits: []CommitNode{
				{Commit: Commit{Message: "Commit 1", Author: GitAuthor{Email: "user@example.com"}}},
			},
			ReviewDecision: "APPROVED",
		}

		if prState.Body != "PR description" {
			t.Errorf("Body = %v, want 'PR description'", prState.Body)
		}
		if len(prState.Comments) != 1 {
			t.Errorf("Comments length = %v, want 1", len(prState.Comments))
		}
		if len(prState.Commits) != 1 {
			t.Errorf("Commits length = %v, want 1", len(prState.Commits))
		}
		if prState.ReviewDecision != "APPROVED" {
			t.Errorf("ReviewDecision = %v, want 'APPROVED'", prState.ReviewDecision)
		}
	})

	t.Run("PRState with empty arrays", func(t *testing.T) {
		prState := PRState{
			Body:           "Empty PR",
			Comments:       []Comment{},
			Commits:        []CommitNode{},
			ReviewDecision: "",
		}

		if len(prState.Comments) != 0 {
			t.Errorf("Expected empty Comments, got length %d", len(prState.Comments))
		}
		if len(prState.Commits) != 0 {
			t.Errorf("Expected empty Commits, got length %d", len(prState.Commits))
		}
	})

	t.Run("PRState with nil arrays", func(t *testing.T) {
		prState := PRState{
			Body:           "PR with nil arrays",
			Comments:       nil,
			Commits:        nil,
			ReviewDecision: "",
		}

		// Nil slices have length 0
		if len(prState.Comments) != 0 {
			t.Errorf("Expected nil Comments to have length 0, got %d", len(prState.Comments))
		}
		if len(prState.Commits) != 0 {
			t.Errorf("Expected nil Commits to have length 0, got %d", len(prState.Commits))
		}

		// Document behavior: nil vs empty slice
		t.Log("NOTE: Nil slices and empty slices both have length 0")
		t.Log("But nil == true for nil slice, false for empty slice")
		t.Log("This can cause issues in comparisons if not careful")
	})
}

// TestComment tests the Comment struct
func TestComment(t *testing.T) {
	t.Run("create Comment with all fields", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		comment := Comment{
			Body:      "This is a comment",
			Author:    Author{Login: "testuser"},
			CreatedAt: testTime,
		}

		if comment.Body != "This is a comment" {
			t.Errorf("Body = %v, want 'This is a comment'", comment.Body)
		}
		if comment.Author.Login != "testuser" {
			t.Errorf("Author.Login = %v, want 'testuser'", comment.Author.Login)
		}
		if !comment.CreatedAt.Equal(testTime) {
			t.Errorf("CreatedAt = %v, want %v", comment.CreatedAt, testTime)
		}
	})

	t.Run("Comment with empty author", func(t *testing.T) {
		comment := Comment{
			Body:      "Anonymous comment",
			Author:    Author{Login: ""},
			CreatedAt: time.Now(),
		}

		if comment.Author.Login != "" {
			t.Errorf("Expected empty author login, got %v", comment.Author.Login)
		}

		t.Log("RISK: Empty author login may cause issues in filtering")
		t.Log("  poller.go uses username to filter self comments")
		t.Log("  Empty username could match unexpectedly")
	})

	t.Run("Comment with multiline body", func(t *testing.T) {
		multilineBody := "Line 1\nLine 2\nLine 3"
		comment := Comment{
			Body:      multilineBody,
			Author:    Author{Login: "user"},
			CreatedAt: time.Now(),
		}

		if comment.Body != multilineBody {
			t.Errorf("Body = %v, want %v", comment.Body, multilineBody)
		}
	})

	t.Run("Comment with special characters", func(t *testing.T) {
		specialBody := "Comment with emoji 🚀 and unicode: こんにちは"
		comment := Comment{
			Body:      specialBody,
			Author:    Author{Login: "user"},
			CreatedAt: time.Now(),
		}

		if comment.Body != specialBody {
			t.Errorf("Body = %v, want %v", comment.Body, specialBody)
		}
	})
}

// TestAuthor tests the Author struct
func TestAuthor(t *testing.T) {
	t.Run("create Author with login", func(t *testing.T) {
		author := Author{Login: "octocat"}

		if author.Login != "octocat" {
			t.Errorf("Login = %v, want 'octocat'", author.Login)
		}
	})

	t.Run("Author with empty login", func(t *testing.T) {
		author := Author{Login: ""}

		if author.Login != "" {
			t.Errorf("Expected empty login, got %v", author.Login)
		}
	})

	t.Run("Author with special characters in login", func(t *testing.T) {
		// GitHub usernames can contain alphanumeric and hyphens
		author := Author{Login: "user-name-123"}

		if author.Login != "user-name-123" {
			t.Errorf("Login = %v, want 'user-name-123'", author.Login)
		}
	})
}

// TestCommit tests the Commit struct
func TestCommit(t *testing.T) {
	t.Run("create Commit with all fields", func(t *testing.T) {
		commit := Commit{
			Message: "feat: add new feature",
			Author:  GitAuthor{Email: "user@example.com"},
		}

		if commit.Message != "feat: add new feature" {
			t.Errorf("Message = %v, want 'feat: add new feature'", commit.Message)
		}
		if commit.Author.Email != "user@example.com" {
			t.Errorf("Author.Email = %v, want 'user@example.com'", commit.Author.Email)
		}
	})

	t.Run("Commit with empty author email", func(t *testing.T) {
		commit := Commit{
			Message: "Commit with no author",
			Author:  GitAuthor{Email: ""},
		}

		if commit.Author.Email != "" {
			t.Errorf("Expected empty email, got %v", commit.Author.Email)
		}

		t.Log("RISK: Empty author email may cause issues in filtering")
		t.Log("  poller.go uses email to filter self commits")
		t.Log("  Empty email could match unexpectedly")
	})

	t.Run("Commit with multiline message", func(t *testing.T) {
		multilineMessage := "Short subject\n\nLonger body\nwith multiple lines"
		commit := Commit{
			Message: multilineMessage,
			Author:  GitAuthor{Email: "user@example.com"},
		}

		if commit.Message != multilineMessage {
			t.Errorf("Message = %v, want %v", commit.Message, multilineMessage)
		}
	})
}

// TestGitAuthor tests the GitAuthor struct
func TestGitAuthor(t *testing.T) {
	t.Run("create GitAuthor with email", func(t *testing.T) {
		author := GitAuthor{Email: "user@example.com"}

		if author.Email != "user@example.com" {
			t.Errorf("Email = %v, want 'user@example.com'", author.Email)
		}
	})

	t.Run("GitAuthor with empty email", func(t *testing.T) {
		author := GitAuthor{Email: ""}

		if author.Email != "" {
			t.Errorf("Expected empty email, got %v", author.Email)
		}
	})

	t.Run("GitAuthor email vs Author login", func(t *testing.T) {
		// Document the difference between comment authors and commit authors
		commentAuthor := Author{Login: "username"}
		commitAuthor := GitAuthor{Email: "username@example.com"}

		t.Logf("Comment author uses Login: %v", commentAuthor.Login)
		t.Logf("Commit author uses Email: %v", commitAuthor.Email)
		
		t.Log("NOTE: Poller uses different fields for filtering:")
		t.Log("  - Comments filtered by Author.Login (username)")
		t.Log("  - Commits filtered by GitAuthor.Email (email)")
		t.Log("  These may not match for the same user!")
	})
}

// TestRegistry tests the Registry type
func TestRegistry(t *testing.T) {
	t.Run("create empty Registry", func(t *testing.T) {
		registry := make(Registry)

		if len(registry) != 0 {
			t.Errorf("Expected empty registry, got length %d", len(registry))
		}
	})

	t.Run("add entries to Registry", func(t *testing.T) {
		registry := make(Registry)
		url := "https://github.com/owner/repo/pull/1"
		state := PRState{
			Body:     "Test PR",
			Comments: []Comment{},
			Commits:  []CommitNode{},
		}

		registry[url] = state

		if len(registry) != 1 {
			t.Errorf("Expected registry length 1, got %d", len(registry))
		}

		storedState, exists := registry[url]
		if !exists {
			t.Error("Expected entry to exist in registry")
		}
		if storedState.Body != "Test PR" {
			t.Errorf("Stored body = %v, want 'Test PR'", storedState.Body)
		}
	})

	t.Run("Registry is reference type", func(t *testing.T) {
		registry1 := make(Registry)
		registry1["key1"] = PRState{Body: "State 1"}

		// Assigning to another variable creates reference, not copy
		registry2 := registry1
		registry2["key2"] = PRState{Body: "State 2"}

		if len(registry1) != 2 {
			t.Errorf("Expected registry1 length 2 (shared reference), got %d", len(registry1))
		}

		t.Log("NOTE: Registry is a map (reference type)")
		t.Log("  Assigning to another variable shares the same underlying map")
		t.Log("  Changes to one affect the other")
	})
}

// TestPRQuery tests the PRQuery struct
func TestPRQuery(t *testing.T) {
	t.Run("create PRQuery with nested structure", func(t *testing.T) {
		testTime := time.Now()
		query := PRQuery{
			Repository: struct {
				PullRequest PRData `graphql:"pullRequest(number: $prNumber)"`
			}{
				PullRequest: PRData{
					Body: "Query test",
					Comments: PRComments{
						Nodes: []Comment{
							{Body: "Comment", Author: Author{Login: "user"}, CreatedAt: testTime},
						},
					},
					Commits: PRCommits{
						Nodes: []CommitNode{
							{Commit: Commit{Message: "Commit"}},
						},
					},
					ReviewDecision: "APPROVED",
				},
			},
		}

		if query.Repository.PullRequest.Body != "Query test" {
			t.Errorf("Body = %v, want 'Query test'", query.Repository.PullRequest.Body)
		}
		if len(query.Repository.PullRequest.Comments.Nodes) != 1 {
			t.Errorf("Comments count = %v, want 1", len(query.Repository.PullRequest.Comments.Nodes))
		}
		if len(query.Repository.PullRequest.Commits.Nodes) != 1 {
			t.Errorf("Commits count = %v, want 1", len(query.Repository.PullRequest.Commits.Nodes))
		}
	})

	t.Run("PRQuery GraphQL tags", func(t *testing.T) {
		// Document the GraphQL structure tags
		t.Log("PRQuery struct tags:")
		t.Log("  Repository: `graphql:\"repository(owner: $owner, name: $repo)\"`")
		t.Log("  PullRequest: `graphql:\"pullRequest(number: $prNumber)\"`")
		t.Log("  Comments: `graphql:\"comments(first: 100)\"`")
		t.Log("  Commits: `graphql:\"commits(first: 100)\"`")
		t.Log("")
		t.Log("LIMITATION: first 100 is hardcoded in struct tags")
		t.Log("  Cannot query more than 100 comments or commits")
		t.Log("  Large PRs will have truncated data")
	})
}

// TestPRData tests the PRData struct
func TestPRData(t *testing.T) {
	t.Run("PRData with 100 comments", func(t *testing.T) {
		comments := make([]Comment, 100)
		for i := 0; i < 100; i++ {
			comments[i] = Comment{Body: "Comment", Author: Author{Login: "user"}, CreatedAt: time.Now()}
		}

		data := PRData{
			Body: "PR with max comments",
			Comments: PRComments{
				Nodes: comments,
			},
			Commits: PRCommits{
				Nodes: []CommitNode{},
			},
		}

		if len(data.Comments.Nodes) != 100 {
			t.Errorf("Comments count = %v, want 100", len(data.Comments.Nodes))
		}
	})

	t.Run("PRData with 100 commits", func(t *testing.T) {
		commits := make([]CommitNode, 100)
		for i := 0; i < 100; i++ {
			commits[i] = CommitNode{Commit: Commit{Message: "Commit"}}
		}

		data := PRData{
			Body:     "PR with max commits",
			Comments: PRComments{Nodes: []Comment{}},
			Commits: PRCommits{
				Nodes: commits,
			},
		}

		if len(data.Commits.Nodes) != 100 {
			t.Errorf("Commits count = %v, want 100", len(data.Commits.Nodes))
		}
	})
}

// TestCommitNode tests the CommitNode wrapper struct
func TestCommitNode(t *testing.T) {
	t.Run("create CommitNode", func(t *testing.T) {
		node := CommitNode{
			Commit: Commit{
				Message: "Test commit",
				Author:  GitAuthor{Email: "test@example.com"},
			},
		}

		if node.Commit.Message != "Test commit" {
			t.Errorf("Message = %v, want 'Test commit'", node.Commit.Message)
		}
	})

	t.Run("CommitNode is a wrapper", func(t *testing.T) {
		t.Log("NOTE: CommitNode wraps Commit")
		t.Log("  This matches GraphQL response structure")
		t.Log("  GitHub API returns commits as { commit: { ... } }")
	})
}
