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
			got := comparePRStates(tt.oldState, tt.newState)
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
			got := compareComments(tt.oldComments, tt.newComments)
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
			got := compareCommits(tt.oldCommits, tt.newCommits)
			if got != tt.want {
				t.Errorf("compareCommits() = %v, want %v", got, tt.want)
			}
		})
	}
}
