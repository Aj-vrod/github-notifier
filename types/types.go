package types

// PRInfo holds the parsed components of a GitHub PR URL
type PRInfo struct {
	Owner  string
	Repo   string
	Number int
}

type PRState struct {
	Exists   bool
	Comments any
	Commits  any
}

type Registry map[string]PRState
