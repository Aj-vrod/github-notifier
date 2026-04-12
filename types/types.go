package types

import "time"

// PRInfo holds the parsed components of a GitHub PR URL
type PRInfo struct {
	URL    string
	Owner  string
	Repo   string
	Number int
}

type PRState struct {
	Body     string
	Comments []Comment
	Commits  []CommitNode
}

type Registry map[string]PRState

type PRQuery struct {
	Repository struct {
		PullRequest PRData `graphql:"pullRequest(number: $prNumber)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

type PRData struct {
	Comments PRComments `graphql:"comments(first: 100)"`
	Commits  PRCommits  `graphql:"commits(first: 100)"`
	Body     string
}

type PRComments struct {
	Nodes []Comment
}

type Comment struct {
	Body      string
	Author    Author
	CreatedAt time.Time
}

type Author struct {
	Login string
}

type PRCommits struct {
	Nodes []CommitNode
}

type CommitNode struct {
	Commit Commit
}

type Commit struct {
	Message string
}
