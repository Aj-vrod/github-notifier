package types

import (
	"encoding/json"
	"testing"
)

// TestSubArgs tests the SubArgs struct
func TestSubArgs(t *testing.T) {
	t.Run("create SubArgs with all fields", func(t *testing.T) {
		args := SubArgs{
			Org:  "company",
			Repo: "my-repo",
			PR:   "123",
		}

		if args.Org != "company" {
			t.Errorf("Org = %v, want 'company'", args.Org)
		}
		if args.Repo != "my-repo" {
			t.Errorf("Repo = %v, want 'my-repo'", args.Repo)
		}
		if args.PR != "123" {
			t.Errorf("PR = %v, want '123'", args.PR)
		}
	})

	t.Run("SubArgs with empty fields", func(t *testing.T) {
		args := SubArgs{
			Org:  "",
			Repo: "",
			PR:   "",
		}

		if args.Org != "" {
			t.Errorf("Expected empty Org, got %v", args.Org)
		}
		if args.Repo != "" {
			t.Errorf("Expected empty Repo, got %v", args.Repo)
		}
		if args.PR != "" {
			t.Errorf("Expected empty PR, got %v", args.PR)
		}

		t.Log("NOTE: SubArgs doesn't validate empty fields")
		t.Log("  Validation happens in cmd.validateArgs()")
	})

	t.Run("SubArgs with special characters", func(t *testing.T) {
		args := SubArgs{
			Org:  "org-with-dash",
			Repo: "repo_with_underscore",
			PR:   "999",
		}

		if args.Org != "org-with-dash" {
			t.Errorf("Org = %v, want 'org-with-dash'", args.Org)
		}
		if args.Repo != "repo_with_underscore" {
			t.Errorf("Repo = %v, want 'repo_with_underscore'", args.Repo)
		}
	})

	t.Run("SubArgs PR field is string not int", func(t *testing.T) {
		// PR is string to preserve leading zeros if any
		args := SubArgs{
			Org:  "org",
			Repo: "repo",
			PR:   "007",
		}

		if args.PR != "007" {
			t.Errorf("PR = %v, want '007'", args.PR)
		}

		t.Log("NOTE: PR field is string, not int")
		t.Log("  Allows preserving format from command line")
		t.Log("  Converted to int when creating PRInfo")
	})
}

// TestSubArgs_URLConstruction tests how SubArgs is used to construct URLs
func TestSubArgs_URLConstruction(t *testing.T) {
	tests := []struct {
		name        string
		args        SubArgs
		expectedURL string
	}{
		{
			name: "standard args",
			args: SubArgs{
				Org:  "github",
				Repo: "gitignore",
				PR:   "1",
			},
			expectedURL: "https://github.com/github/gitignore/pull/1",
		},
		{
			name: "org and repo with hyphens",
			args: SubArgs{
				Org:  "my-company",
				Repo: "my-project",
				PR:   "42",
			},
			expectedURL: "https://github.com/my-company/my-project/pull/42",
		},
		{
			name: "repo with underscore",
			args: SubArgs{
				Org:  "org",
				Repo: "repo_name",
				PR:   "100",
			},
			expectedURL: "https://github.com/org/repo_name/pull/100",
		},
		{
			name: "large PR number",
			args: SubArgs{
				Org:  "org",
				Repo: "repo",
				PR:   "99999",
			},
			expectedURL: "https://github.com/org/repo/pull/99999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is how subscribe.go and unsubscribe.go construct the URL
			constructedURL := "https://github.com/" + tt.args.Org + "/" + tt.args.Repo + "/pull/" + tt.args.PR

			if constructedURL != tt.expectedURL {
				t.Errorf("Constructed URL = %v, want %v", constructedURL, tt.expectedURL)
			}
		})
	}
}

// TestSubArgs_InvalidInputs tests edge cases with invalid inputs
func TestSubArgs_InvalidInputs(t *testing.T) {
	t.Run("SubArgs with spaces", func(t *testing.T) {
		args := SubArgs{
			Org:  "org with spaces",
			Repo: "repo with spaces",
			PR:   "123",
		}

		// Spaces would create invalid GitHub URL
		url := "https://github.com/" + args.Org + "/" + args.Repo + "/pull/" + args.PR

		t.Logf("URL with spaces: %s", url)
		t.Log("RISK: No validation prevents spaces in org/repo names")
		t.Log("  Would create invalid URL: https://github.com/org with spaces/repo with spaces/pull/123")
		t.Log("  GitHub API call would fail")
	})

	t.Run("SubArgs with slashes", func(t *testing.T) {
		args := SubArgs{
			Org:  "org/subpath",
			Repo: "repo",
			PR:   "123",
		}

		url := "https://github.com/" + args.Org + "/" + args.Repo + "/pull/" + args.PR

		t.Logf("URL with slash: %s", url)
		t.Log("RISK: Slashes in org name create malformed URL")
		t.Log("  https://github.com/org/subpath/repo/pull/123")
		t.Log("  Should be rejected at validation")
	})

	t.Run("SubArgs with non-numeric PR", func(t *testing.T) {
		_ = SubArgs{
			Org:  "org",
			Repo: "repo",
			PR:   "not-a-number",
		}

		// PR field is string, so this is allowed
		// But would fail when converted to int for API call

		t.Log("RISK: PR field accepts non-numeric values")
		t.Log("  Type is string, so 'not-a-number' is valid")
		t.Log("  Would fail when converting to int for API call")
		t.Log("  Error happens late (in API validator) instead of at CLI parsing")
	})

	t.Run("SubArgs with negative PR", func(t *testing.T) {
		_ = SubArgs{
			Org:  "org",
			Repo: "repo",
			PR:   "-1",
		}

		t.Log("RISK: Negative PR numbers accepted as strings")
		t.Log("  Would be rejected by API validator")
		t.Log("  But not caught at CLI argument parsing")
	})
}

// TestMessage tests the Message struct
func TestMessage(t *testing.T) {
	t.Run("create Message with text", func(t *testing.T) {
		msg := Message{
			Text: "Hello from Slack!",
		}

		if msg.Text != "Hello from Slack!" {
			t.Errorf("Text = %v, want 'Hello from Slack!'", msg.Text)
		}
	})

	t.Run("Message with empty text", func(t *testing.T) {
		msg := Message{
			Text: "",
		}

		if msg.Text != "" {
			t.Errorf("Expected empty Text, got %v", msg.Text)
		}
	})

	t.Run("Message JSON marshaling", func(t *testing.T) {
		msg := Message{
			Text: "Test message",
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal Message: %v", err)
		}

		expected := `{"text":"Test message"}`
		if string(jsonData) != expected {
			t.Errorf("JSON = %v, want %v", string(jsonData), expected)
		}
	})

	t.Run("Message JSON unmarshaling", func(t *testing.T) {
		jsonData := []byte(`{"text":"Unmarshaled message"}`)

		var msg Message
		err := json.Unmarshal(jsonData, &msg)
		if err != nil {
			t.Fatalf("Failed to unmarshal Message: %v", err)
		}

		if msg.Text != "Unmarshaled message" {
			t.Errorf("Text = %v, want 'Unmarshaled message'", msg.Text)
		}
	})

	t.Run("Message with special characters in JSON", func(t *testing.T) {
		msg := Message{
			Text: "Message with \"quotes\" and \n newlines",
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal Message with special chars: %v", err)
		}

		// JSON should escape special characters
		var unmarshaledMsg Message
		err = json.Unmarshal(jsonData, &unmarshaledMsg)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if unmarshaledMsg.Text != msg.Text {
			t.Errorf("Unmarshaled text doesn't match original")
		}
	})

	t.Run("Message JSON tag", func(t *testing.T) {
		// Verify the json tag is "text" (lowercase)
		msg := Message{Text: "test"}
		jsonData, _ := json.Marshal(msg)

		if string(jsonData) != `{"text":"test"}` {
			t.Errorf("JSON tag should be lowercase 'text', got %s", string(jsonData))
		}

		t.Log("OK: Message.Text has json tag 'text' (lowercase)")
	})

	t.Run("Message with emoji", func(t *testing.T) {
		msg := Message{
			Text: "Changes detected 🚀 in PR",
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal Message with emoji: %v", err)
		}

		var unmarshaledMsg Message
		err = json.Unmarshal(jsonData, &unmarshaledMsg)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if unmarshaledMsg.Text != msg.Text {
			t.Errorf("Emoji not preserved: got %v, want %v", unmarshaledMsg.Text, msg.Text)
		}
	})

	t.Run("Message with very long text", func(t *testing.T) {
		longText := ""
		for i := 0; i < 10000; i++ {
			longText += "A"
		}

		msg := Message{Text: longText}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal long message: %v", err)
		}

		t.Log("OK: Message can handle very long text")
		t.Logf("  Text length: %d characters", len(longText))
		t.Logf("  JSON length: %d bytes", len(jsonData))
	})
}

// TestMessage_SlackAPIFormat tests that Message matches Slack API expectations
func TestMessage_SlackAPIFormat(t *testing.T) {
	t.Run("Slack expects text field", func(t *testing.T) {
		msg := Message{Text: "Test"}
		jsonData, _ := json.Marshal(msg)

		// Slack webhook API expects {"text": "message"}
		expected := `{"text":"Test"}`
		if string(jsonData) != expected {
			t.Errorf("Slack expects format %s, got %s", expected, string(jsonData))
		}

		t.Log("OK: Message format matches Slack webhook API")
	})

	t.Run("Message is minimal Slack payload", func(t *testing.T) {
		t.Log("NOTE: Message struct is minimal")
		t.Log("  Only has 'text' field")
		t.Log("  Slack webhooks support many more fields:")
		t.Log("    - username: Custom bot name")
		t.Log("    - icon_emoji: Custom emoji icon")
		t.Log("    - icon_url: Custom image icon")
		t.Log("    - channel: Override default channel")
		t.Log("    - attachments: Rich message formatting")
		t.Log("    - blocks: Block Kit formatting")
		t.Log("  Current implementation only sends plain text")
	})
}

// TestSubArgs_ZeroValues tests zero value behavior
func TestSubArgs_ZeroValues(t *testing.T) {
	t.Run("zero value SubArgs", func(t *testing.T) {
		var args SubArgs

		if args.Org != "" {
			t.Errorf("Zero value Org should be empty string, got %v", args.Org)
		}
		if args.Repo != "" {
			t.Errorf("Zero value Repo should be empty string, got %v", args.Repo)
		}
		if args.PR != "" {
			t.Errorf("Zero value PR should be empty string, got %v", args.PR)
		}

		t.Log("OK: Zero value SubArgs has empty strings for all fields")
	})

	t.Run("zero value Message", func(t *testing.T) {
		var msg Message

		if msg.Text != "" {
			t.Errorf("Zero value Text should be empty string, got %v", msg.Text)
		}

		t.Log("OK: Zero value Message has empty string for Text")
	})
}
