package poller

import (
	"Aj-vrod/github-notifier/types"
	"testing"
	"time"
)

func TestComparePRStates(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name     string
		oldState types.PRState
		newState types.PRState
		want     bool
	}{
		{
			name: "no changes",
			oldState: types.PRState{
				Body: "same body",
				Comments: []types.Comment{
					{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
				},
			},
			newState: types.PRState{
				Body: "same body",
				Comments: []types.Comment{
					{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
				},
			},
			want: false,
		},
		{
			name: "body changed",
			oldState: types.PRState{
				Body:     "old body",
				Comments: []types.Comment{},
				Commits:  []types.CommitNode{},
			},
			newState: types.PRState{
				Body:     "new body",
				Comments: []types.Comment{},
				Commits:  []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "comment count increased",
			oldState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
					{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "comment body changed",
			oldState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "old comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "new comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "comment author changed",
			oldState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "comment timestamp changed",
			oldState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime.Add(1 * time.Hour)},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "commit count increased",
			oldState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
				},
			},
			newState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
					{Commit: types.Commit{Message: "commit2"}},
				},
			},
			want: true,
		},
		{
			name: "commit count decreased",
			oldState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
					{Commit: types.Commit{Message: "commit2"}},
				},
			},
			newState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "commit1"}},
				},
			},
			want: true,
		},
		{
			name: "multiple changes - body and comments",
			oldState: types.PRState{
				Body: "old body",
				Comments: []types.Comment{
					{Body: "old comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "new body",
				Comments: []types.Comment{
					{Body: "new comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "empty to non-empty comments",
			oldState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits:  []types.CommitNode{},
			},
			newState: types.PRState{
				Body: "body",
				Comments: []types.Comment{
					{Body: "first comment", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				},
				Commits: []types.CommitNode{},
			},
			want: true,
		},
		{
			name: "empty to non-empty commits",
			oldState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits:  []types.CommitNode{},
			},
			newState: types.PRState{
				Body:     "body",
				Comments: []types.Comment{},
				Commits: []types.CommitNode{
					{Commit: types.Commit{Message: "first commit"}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{User: "testuser", UserEmail: "test@example.com"}
			got := comparePRStates(tt.oldState, tt.newState, cfg)
			if got != tt.want {
				t.Errorf("comparePRStates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareComments(t *testing.T) {
	testTime := time.Now()
	laterTime := testTime.Add(1 * time.Hour)

	tests := []struct {
		name        string
		oldComments []types.Comment
		newComments []types.Comment
		want        bool
	}{
		{
			name:        "both empty",
			oldComments: []types.Comment{},
			newComments: []types.Comment{},
			want:        false,
		},
		{
			name:        "same single comment",
			oldComments: []types.Comment{{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime}},
			newComments: []types.Comment{{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime}},
			want:        false,
		},
		{
			name:        "different count - old empty",
			oldComments: []types.Comment{},
			newComments: []types.Comment{{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime}},
			want:        true,
		},
		{
			name:        "different count - new empty",
			oldComments: []types.Comment{{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime}},
			newComments: []types.Comment{},
			want:        true,
		},
		{
			name: "different count - multiple comments",
			oldComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			want: true,
		},
		{
			name: "different body - same count",
			oldComments: []types.Comment{
				{Body: "old body", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "new body", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			},
			want: true,
		},
		{
			name: "different author - same count",
			oldComments: []types.Comment{
				{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "test", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			want: true,
		},
		{
			name: "different timestamp - same count",
			oldComments: []types.Comment{
				{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "test", Author: types.Author{Login: "user1"}, CreatedAt: laterTime},
			},
			want: true,
		},
		{
			name: "multiple comments - no changes",
			oldComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			want: false,
		},
		{
			name: "multiple comments - second one changed",
			oldComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				{Body: "old comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			newComments: []types.Comment{
				{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
				{Body: "new comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareComments(tt.oldComments, tt.newComments, "testuser")
			if got != tt.want {
				t.Errorf("compareComments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareCommits(t *testing.T) {
	tests := []struct {
		name       string
		oldCommits []types.CommitNode
		newCommits []types.CommitNode
		want       bool
	}{
		{
			name:       "both empty",
			oldCommits: []types.CommitNode{},
			newCommits: []types.CommitNode{},
			want:       false,
		},
		{
			name: "same count - single commit",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			want: false,
		},
		{
			name: "same count - multiple commits",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
				{Commit: types.Commit{Message: "commit2"}},
			},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
				{Commit: types.Commit{Message: "commit2"}},
			},
			want: false,
		},
		{
			name:       "different count - old empty",
			oldCommits: []types.CommitNode{},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			want: true,
		},
		{
			name: "different count - new empty",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			newCommits: []types.CommitNode{},
			want:       true,
		},
		{
			name: "count increased",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
				{Commit: types.Commit{Message: "commit2"}},
			},
			want: true,
		},
		{
			name: "count decreased",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
				{Commit: types.Commit{Message: "commit2"}},
			},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "commit1"}},
			},
			want: true,
		},
		{
			name: "same count but different messages - not detected",
			oldCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "old message"}},
			},
			newCommits: []types.CommitNode{
				{Commit: types.Commit{Message: "new message"}},
			},
			want: false, // Current implementation only checks count, not content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareCommits(tt.oldCommits, tt.newCommits, "test@example.com")
			if got != tt.want {
				t.Errorf("compareCommits() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BUG TESTS: The following tests document array bounds bugs in compareComments and compareCommits.
// These bugs occur at internal/poller/poller.go lines 125 and 142.
// The functions iterate over oldComments/oldCommits and access newComments[i]/newCommits[i]
// without checking if i < len(newComments/newCommits), causing panics when the new array is shorter.

// TestCompareComments_ArrayShrinkagePanic tests the array bounds bug in compareComments.
// BUG: When oldComments has MORE elements than newComments, the loop at line 125
// accesses newComments[i] without bounds checking, causing an index out of range panic.
// This can happen if comments are deleted from a PR.
func TestCompareComments_ArrayShrinkagePanic(t *testing.T) {
	testTime := time.Now()

	t.Run("old has 5 comments, new has 3 - WILL PANIC", func(t *testing.T) {
		// Skip this test by default as it will panic
		// Run with: go test -v -run TestCompareComments_ArrayShrinkagePanic/old
		if testing.Short() {
			t.Skip("Skipping panic test in short mode")
		}

		oldComments := []types.Comment{
			{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			{Body: "comment3", Author: types.Author{Login: "user3"}, CreatedAt: testTime},
			{Body: "comment4", Author: types.Author{Login: "user4"}, CreatedAt: testTime},
			{Body: "comment5", Author: types.Author{Login: "user5"}, CreatedAt: testTime},
		}
		newComments := []types.Comment{
			{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			{Body: "comment2", Author: types.Author{Login: "user2"}, CreatedAt: testTime},
			{Body: "comment3", Author: types.Author{Login: "user3"}, CreatedAt: testTime},
		}

		// This should panic with "index out of range [3] with length 3"
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic occurred: %v", r)
				t.Logf("BUG: compareComments panics when oldComments is longer than newComments")
			} else {
				t.Error("Expected panic did not occur - bug may have been fixed")
			}
		}()

		compareComments(oldComments, newComments, "testuser")
	})

	t.Run("old has 1 comment, new is empty - WILL PANIC", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping panic test in short mode")
		}

		oldComments := []types.Comment{
			{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
		}
		newComments := []types.Comment{}

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic occurred: %v", r)
				t.Logf("BUG: compareComments panics when newComments is empty but oldComments is not")
			} else {
				t.Error("Expected panic did not occur - bug may have been fixed")
			}
		}()

		compareComments(oldComments, newComments, "testuser")
	})
}

// TestCompareComments_EdgeCases tests edge cases that should NOT panic but may have unexpected behavior.
func TestCompareComments_EdgeCases(t *testing.T) {
	testTime := time.Now()

	t.Run("new comment added by self - should be filtered", func(t *testing.T) {
		oldComments := []types.Comment{
			{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
		}
		newComments := []types.Comment{
			{Body: "comment1", Author: types.Author{Login: "user1"}, CreatedAt: testTime},
			{Body: "my own comment", Author: types.Author{Login: "testuser"}, CreatedAt: testTime},
		}

		// Should return false because the new comment is by the user themselves
		got := compareComments(oldComments, newComments, "testuser")
		if got {
			t.Error("Expected false (self comment should be filtered), got true")
		}
	})

	t.Run("all new comments by self", func(t *testing.T) {
		newComments := []types.Comment{
			{Body: "my comment 1", Author: types.Author{Login: "testuser"}, CreatedAt: testTime},
			{Body: "my comment 2", Author: types.Author{Login: "testuser"}, CreatedAt: testTime},
		}

		// Should return false because all comments are by the user themselves
		got := compareComments(newComments[:1], newComments, "testuser")
		if got {
			t.Error("Expected false (all new comments by self should be filtered), got true")
		}
	})

	t.Run("empty arrays should not panic", func(t *testing.T) {
		oldComments := []types.Comment{}
		newComments := []types.Comment{}

		// Should not panic and should return false
		got := compareComments(oldComments, newComments, "testuser")
		if got {
			t.Error("Expected false (both empty), got true")
		}
	})
}

// TestCompareCommits_ArrayShrinkagePanic tests the array bounds bug in compareCommits.
// BUG: Although compareCommits doesn't iterate through commits like compareComments does,
// it could still have similar issues if the implementation is expanded in the future.
// Currently it only checks count changes, but documenting potential risk.
func TestCompareCommits_ArrayShrinkagePanic(t *testing.T) {
	t.Run("commits array shrinkage is detected", func(t *testing.T) {
		oldCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "commit1", Author: types.GitAuthor{Email: "user1@example.com"}}},
			{Commit: types.Commit{Message: "commit2", Author: types.GitAuthor{Email: "user2@example.com"}}},
			{Commit: types.Commit{Message: "commit3", Author: types.GitAuthor{Email: "user3@example.com"}}},
		}
		newCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "commit1", Author: types.GitAuthor{Email: "user1@example.com"}}},
		}

		// This should return true (count decreased) and should NOT panic
		got := compareCommits(oldCommits, newCommits, "testuser@example.com")
		if !got {
			t.Error("Expected true (count decreased), got false")
		}
	})

	t.Run("new commit added by self - should be filtered", func(t *testing.T) {
		oldCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "commit1", Author: types.GitAuthor{Email: "user1@example.com"}}},
		}
		newCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "commit1", Author: types.GitAuthor{Email: "user1@example.com"}}},
			{Commit: types.Commit{Message: "my commit", Author: types.GitAuthor{Email: "testuser@example.com"}}},
		}

		// Should return false because the new commit is by the user themselves
		got := compareCommits(oldCommits, newCommits, "testuser@example.com")
		if got {
			t.Error("Expected false (self commit should be filtered), got true")
		}
	})
}

// TestCompareCommits_MissingContentComparison documents that compareCommits only checks
// count changes, not actual commit content changes. This is different from compareComments.
func TestCompareCommits_MissingContentComparison(t *testing.T) {
	t.Run("commit message changed but count same - NOT DETECTED", func(t *testing.T) {
		oldCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "original message", Author: types.GitAuthor{Email: "user@example.com"}}},
		}
		newCommits := []types.CommitNode{
			{Commit: types.Commit{Message: "amended message", Author: types.GitAuthor{Email: "user@example.com"}}},
		}

		// BUG/LIMITATION: compareCommits doesn't check message content, only count
		got := compareCommits(oldCommits, newCommits, "other@example.com")
		if got {
			t.Error("Expected false (current implementation doesn't check content), got true")
		}
	})
}
