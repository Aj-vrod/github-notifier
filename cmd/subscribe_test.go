package cmd

import (
	"Aj-vrod/github-notifier/types"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSubscribeCmd_Configuration(t *testing.T) {
	if subscribeCmd.Use != "sub" {
		t.Errorf("Expected Use='sub', got '%s'", subscribeCmd.Use)
	}
	if subscribeCmd.Short != "Subscribe a PR" {
		t.Errorf("Expected Short='Subscribe a PR', got '%s'", subscribeCmd.Short)
	}
	if len(subscribeCmd.ValidArgs) != 3 {
		t.Errorf("Expected 3 valid args, got %d", len(subscribeCmd.ValidArgs))
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid args",
			args:    []string{"org=company", "repo=my-repo", "pr=123"},
			wantErr: false,
		},
		{
			name:    "missing org",
			args:    []string{"repo=my-repo", "pr=123"},
			wantErr: true,
			errMsg:  "missing org argument",
		},
		{
			name:    "missing repo",
			args:    []string{"org=company", "pr=123"},
			wantErr: true,
			errMsg:  "missing repo argument",
		},
		{
			name:    "missing pr",
			args:    []string{"org=company", "repo=my-repo"},
			wantErr: true,
			errMsg:  "missing pr argument",
		},
		{
			name:    "all missing",
			args:    []string{},
			wantErr: true,
			errMsg:  "missing org argument",
		},
		{
			name:    "valid with extra args",
			args:    []string{"org=company", "repo=my-repo", "pr=123", "extra=value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error '%s', got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%s'", err)
				}
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOrg  string
		wantRepo string
		wantPR   string
	}{
		{
			name:     "standard args",
			args:     []string{"org=company", "repo=my-repo", "pr=123"},
			wantOrg:  "company",
			wantRepo: "my-repo",
			wantPR:   "123",
		},
		{
			name:     "different order",
			args:     []string{"pr=456", "org=test-org", "repo=test-repo"},
			wantOrg:  "test-org",
			wantRepo: "test-repo",
			wantPR:   "456",
		},
		{
			name:     "with hyphens and numbers",
			args:     []string{"org=my-org-123", "repo=repo-name-v2", "pr=999"},
			wantOrg:  "my-org-123",
			wantRepo: "repo-name-v2",
			wantPR:   "999",
		},
		{
			name:     "empty values",
			args:     []string{"org=", "repo=", "pr="},
			wantOrg:  "",
			wantRepo: "",
			wantPR:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqArgs *types.SubArgs
			parseArgs(tt.args, &reqArgs)

			if reqArgs.Org != tt.wantOrg {
				t.Errorf("Org = %v, want %v", reqArgs.Org, tt.wantOrg)
			}
			if reqArgs.Repo != tt.wantRepo {
				t.Errorf("Repo = %v, want %v", reqArgs.Repo, tt.wantRepo)
			}
			if reqArgs.PR != tt.wantPR {
				t.Errorf("PR = %v, want %v", reqArgs.PR, tt.wantPR)
			}
		})
	}
}

func TestSubscribe_HTTPRequest(t *testing.T) {
	tests := []struct {
		name           string
		args           *types.SubArgs
		serverResponse int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful subscribe",
			args: &types.SubArgs{
				Org:  "company",
				Repo: "my-repo",
				PR:   "123",
			},
			serverResponse: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "server returns 400",
			args: &types.SubArgs{
				Org:  "company",
				Repo: "my-repo",
				PR:   "invalid",
			},
			serverResponse: http.StatusBadRequest,
			wantErr:        true,
			errContains:    "record was not created",
		},
		{
			name: "server returns 500",
			args: &types.SubArgs{
				Org:  "company",
				Repo: "my-repo",
				PR:   "123",
			},
			serverResponse: http.StatusInternalServerError,
			wantErr:        true,
			errContains:    "record was not created",
		},
		{
			name: "server returns 409 conflict",
			args: &types.SubArgs{
				Org:  "company",
				Repo: "my-repo",
				PR:   "123",
			},
			serverResponse: http.StatusConflict,
			wantErr:        true,
			errContains:    "record was not created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify URL path
				if !strings.HasSuffix(r.URL.Path, "/api/v1/subscribe") {
					t.Errorf("Expected path to end with '/api/v1/subscribe', got %s", r.URL.Path)
				}

				// Verify request body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}
				expectedBody := fmt.Sprintf(`{"pr_url": "https://github.com/%s/%s/pull/%s"}`, tt.args.Org, tt.args.Repo, tt.args.PR)
				if string(body) != expectedBody {
					t.Errorf("Expected body %s, got %s", expectedBody, string(body))
				}

				// Send response
				w.WriteHeader(tt.serverResponse)
			}))
			defer server.Close()

			// Temporarily override the default HTTP client to use test server
			// Note: The actual implementation uses a hardcoded localhost:8001
			// This test demonstrates what SHOULD be tested, but the actual function
			// cannot be tested without refactoring to accept a configurable URL

			// BUG DOCUMENTATION: subscribe() function uses hardcoded localhost:8001
			// making it impossible to test without a running server
			t.Logf("BUG: subscribe() uses hardcoded URL 'http://localhost:%d', cannot inject test server", ServerPort)
		})
	}
}

// TestSubscribe_ErrorHandling tests error scenarios in the subscribe function
func TestSubscribe_ErrorHandling(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		_ = &types.SubArgs{
			Org:  "",
			Repo: "",
			PR:   "",
		}

		// This will try to connect to localhost:8001
		// BUG: Cannot test without refactoring to accept configurable URL
		t.Log("BUG: Cannot test error handling without dependency injection for HTTP client/URL")

		// Document expected behavior
		t.Log("Expected: Should validate empty args before making HTTP request")
		t.Log("Actual: No validation, will attempt HTTP request with empty values")

		// Skip actual call as it would fail with connection refused
		t.Skip("Skipping actual call - would require running server")
	})

	t.Run("special characters in args", func(t *testing.T) {
		_ = &types.SubArgs{
			Org:  "org/with/slashes",
			Repo: "repo with spaces",
			PR:   "pr#123",
		}

		t.Log("Expected: Should sanitize or validate special characters")
		t.Log("Actual: No validation, will construct potentially invalid URL")
		t.Skip("Skipping actual call - would require running server")
	})
}

// TestSubscribeCmd_Integration tests the full command execution
func TestSubscribeCmd_Integration(t *testing.T) {
	t.Run("command with valid args", func(t *testing.T) {
		// Create a new command for testing
		cmd := &cobra.Command{}
		cmd.AddCommand(subscribeCmd)

		// Set args
		subscribeCmd.SetArgs([]string{"org=test", "repo=repo", "pr=1"})

		// PreRunE validation should pass
		err := validateArgs([]string{"org=test", "repo=repo", "pr=1"})
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}

		// Note: Cannot test full RunE execution without a running server
		t.Log("BUG: Cannot test full command execution without dependency injection")
	})

	t.Run("command with invalid args", func(t *testing.T) {
		// Missing 'pr' argument
		err := validateArgs([]string{"org=test", "repo=repo"})
		if err == nil {
			t.Error("Expected validation error for missing 'pr' argument")
		}
	})
}

// TestSubscribe_ErrorMessages documents error message quality issues
func TestSubscribe_ErrorMessages(t *testing.T) {
	t.Run("connection refused error message", func(t *testing.T) {
		// BUG: The error message "could not execute the request" is too generic
		// It doesn't distinguish between:
		// - Connection refused (server not running)
		// - Network timeout
		// - DNS resolution failure
		// - TLS handshake failure

		t.Log("BUG: Error message 'could not execute the request' is too generic")
		t.Log("Should provide: 'could not connect to server at localhost:8001: connection refused'")
	})

	t.Run("non-201 status error message", func(t *testing.T) {
		// BUG: Error message references '%s' with err parameter
		// But err is nil when status is not 201 - would print "<nil>"
		// Line 82: return fmt.Errorf("record was not created: %s", err)

		t.Log("BUG: Error message format uses placeholder with nil err, will print 'record was not created: <nil>'")
		t.Log("Should be: fmt.Errorf with res.StatusCode instead")
	})
}
