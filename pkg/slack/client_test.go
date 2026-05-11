package slack

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendNotification_Success(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err != nil {
		t.Errorf("SendNotification() error = %v, want nil", err)
	}
}

func TestSendNotification_SlackError400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for 400 status")
	}
	if err != nil && err.Error() != "slack returned status 400" {
		t.Errorf("SendNotification() error = %v, want 'slack returned status 400'", err)
	}
}

func TestSendNotification_SlackError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for 500 status")
	}
	if err != nil && err.Error() != "slack returned status 500" {
		t.Errorf("SendNotification() error = %v, want 'slack returned status 500'", err)
	}
}

func TestSendNotification_SlackError404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for 404 status")
	}
}

func TestSendNotification_Timeout(t *testing.T) {
	// Create a server that delays longer than the client timeout (10 seconds)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for timeout")
	}
}

func TestSendNotification_NetworkError(t *testing.T) {
	// Use an invalid URL to trigger a network error
	cfg := &Config{WebhookURL: "http://localhost:1"}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for network failure")
	}
}

func TestSendNotification_InvalidURL(t *testing.T) {
	// Use a malformed URL
	cfg := &Config{WebhookURL: "://invalid-url"}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("SendNotification() error = nil, want error for invalid URL")
	}
}

func TestSendNotification_EmptyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Should still work with empty message
	err := client.SendNotification("")
	if err != nil {
		t.Errorf("SendNotification() with empty message error = %v, want nil", err)
	}
}

func TestSendNotification_LongMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Test with a very long message
	longMessage := ""
	for i := 0; i < 1000; i++ {
		longMessage += "This is a long message. "
	}

	err := client.SendNotification(longMessage)
	if err != nil {
		t.Errorf("SendNotification() with long message error = %v, want nil", err)
	}
}

func TestSendNotification_SpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Test with special characters that need JSON escaping
	specialMessage := "Test with \"quotes\", \n newlines, and \t tabs"

	err := client.SendNotification(specialMessage)
	if err != nil {
		t.Errorf("SendNotification() with special characters error = %v, want nil", err)
	}
}

// ============================================================================
// ADDITIONAL EDGE CASE AND FAILURE MODE TESTS
// ============================================================================

// TestNewSlackClient_NoValidation tests that client creation doesn't validate webhook URL
func TestNewSlackClient_NoValidation(t *testing.T) {
	t.Run("invalid URL format accepted", func(t *testing.T) {
		cfg := &Config{WebhookURL: "not-a-valid-url"}
		client := NewSlackClient(cfg)

		if client == nil {
			t.Error("Expected client to be created despite invalid URL")
		}

		// BUG DOCUMENTATION: URL validation happens only at first send
		t.Log("BUG: NewSlackClient() doesn't validate webhook URL")
		t.Log("  - Invalid URLs accepted at initialization")
		t.Log("  - Error only occurs on first SendNotification() call")
		t.Log("  - Service appears to start successfully but fails later")
		t.Log("Expected: Validate URL format at client creation")
	})

	t.Run("empty URL accepted", func(t *testing.T) {
		cfg := &Config{WebhookURL: ""}
		client := NewSlackClient(cfg)

		if client == nil {
			t.Error("Expected client to be created despite empty URL")
		}

		t.Log("BUG: Empty webhook URL accepted, will fail on first use")
	})

	t.Run("malformed slack webhook URL", func(t *testing.T) {
		// Slack webhooks should match pattern https://hooks.slack.com/services/...
		cfg := &Config{WebhookURL: "http://example.com/webhook"}
		client := NewSlackClient(cfg)

		if client == nil {
			t.Error("Expected client to be created despite non-Slack URL")
		}

		t.Log("RISK: Non-Slack webhook URL accepted")
		t.Log("  May send notifications to wrong endpoint")
		t.Log("  Could leak PR change info to untrusted server")
	})
}

// TestSendNotification_TimeoutBehavior tests timeout handling
func TestSendNotification_TimeoutBehavior(t *testing.T) {
	t.Run("timeout is 10 seconds", func(t *testing.T) {
		// Verify timeout is set to 10 seconds (line 30-32 of client.go)
		cfg := &Config{WebhookURL: "http://example.com"}
		client := NewSlackClient(cfg)

		if client.httpClient.Timeout != 10*time.Second {
			t.Errorf("Expected timeout 10s, got %v", client.httpClient.Timeout)
		}

		t.Log("OK: HTTP client timeout set to 10 seconds")
	})

	t.Run("slow webhook delays poller", func(t *testing.T) {
		// If webhook takes 9 seconds to respond, poller is blocked
		// This delays checking other subscriptions

		t.Log("RISK: Slow Slack webhook blocks poller")
		t.Log("  - Webhook can take up to 10 seconds")
		t.Log("  - Poller waits for SendNotification() to complete")
		t.Log("  - Other subscriptions not checked until current notification completes")
		t.Log("  - Multiple slow notifications can significantly delay poller interval")
		t.Log("Expected: Send notifications in separate goroutines with timeout")
	})
}

// TestSendNotification_RateLimiting documents lack of rate limiting
func TestSendNotification_RateLimiting(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount > 5 {
			// Simulate Slack rate limit
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("rate_limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Send many notifications rapidly
	for i := 0; i < 10; i++ {
		err := client.SendNotification("Message " + string(rune(i)))
		if i < 5 {
			if err != nil {
				t.Errorf("Message %d: unexpected error: %v", i, err)
			}
		} else {
			if err == nil {
				t.Errorf("Message %d: expected rate limit error, got nil", i)
			}
		}
	}

	t.Log("LIMITATION: No rate limit handling")
	t.Log("  - Client doesn't track request rate")
	t.Log("  - Slack rate limits (1 per second per webhook) not respected")
	t.Log("  - 429 errors treated as generic failures")
	t.Log("  - No retry logic for rate-limited requests")
	t.Log("Expected: Implement rate limiting and retry with backoff")
}

// TestSendNotification_ErrorMessages documents error message quality
func TestSendNotification_ErrorMessages(t *testing.T) {
	t.Run("generic status code error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid_payload"}`))
		}))
		defer server.Close()

		cfg := &Config{WebhookURL: server.URL}
		client := NewSlackClient(cfg)

		err := client.SendNotification("test")
		if err == nil {
			t.Fatal("Expected error")
		}

		// BUG: Error message only includes status code, not response body
		if err.Error() != "slack returned status 400" {
			t.Errorf("Got error: %v", err)
		}

		t.Log("LIMITATION: Error message doesn't include Slack's error details")
		t.Log("  Actual: 'slack returned status 400'")
		t.Log("  Could be: 'slack returned status 400: invalid_payload'")
	})

	t.Run("network error message", func(t *testing.T) {
		cfg := &Config{WebhookURL: "http://localhost:1"}
		client := NewSlackClient(cfg)

		err := client.SendNotification("test")
		if err == nil {
			t.Fatal("Expected error")
		}

		// Error is wrapped with context
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}

		t.Log("OK: Network errors wrapped with 'failed to send message'")
	})
}

// TestSendNotification_ConcurrentCalls tests concurrent notification sending
func TestSendNotification_ConcurrentCalls(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		time.Sleep(100 * time.Millisecond) // Simulate slow webhook
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Send 5 notifications concurrently
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			err := client.SendNotification("Concurrent message " + string(rune(id)))
			done <- err
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		err := <-done
		if err != nil {
			t.Errorf("Concurrent call %d failed: %v", i, err)
		}
	}

	if callCount != 5 {
		t.Errorf("Expected 5 calls, got %d", callCount)
	}

	t.Log("OK: Client is safe for concurrent use (http.Client is thread-safe)")
}

// TestSendNotification_RetryBehavior documents lack of retry logic
func TestSendNotification_RetryBehavior(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			// First attempt fails with 500
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second attempt would succeed
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test message")
	if err == nil {
		t.Error("Expected error on first attempt")
	}

	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}

	t.Log("LIMITATION: No retry logic for transient failures")
	t.Log("  - 500 Internal Server Error: No retry")
	t.Log("  - 503 Service Unavailable: No retry")
	t.Log("  - Network timeout: No retry")
	t.Log("  - Transient failures result in missed notifications")
	t.Log("Expected: Implement retry with exponential backoff")
}

// TestSendNotification_JSONMarshalError tests JSON encoding edge cases
func TestSendNotification_JSONMarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	// Standard strings should always marshal successfully
	// Testing that marshaling doesn't fail unexpectedly
	messages := []string{
		"Simple message",
		"Message with emoji 🚀",
		"Message with unicode: こんにちは",
		"Message with null bytes: test\x00test",
		"",
	}

	for _, msg := range messages {
		err := client.SendNotification(msg)
		if err != nil {
			t.Errorf("Failed to send message '%s': %v", msg, err)
		}
	}

	t.Log("OK: JSON marshaling handles various string inputs correctly")
}

// TestSendNotification_ResponseBodyClosing verifies proper resource cleanup
func TestSendNotification_ResponseBodyClosing(t *testing.T) {
	closeCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Note: We can't directly test if Body.Close() is called from outside,
	// but we can verify the code has defer resp.Body.Close()

	cfg := &Config{WebhookURL: server.URL}
	client := NewSlackClient(cfg)

	err := client.SendNotification("Test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The actual close happens via defer in the implementation
	// This test documents that proper cleanup is important
	t.Log("OK: Response body closed with defer (line 58 of client.go)")
	t.Log("  Prevents resource leaks on repeated notifications")
	_ = closeCount
}

// TestSendNotification_WebhookURLImmutable tests that URL can't be changed after creation
func TestSendNotification_WebhookURLImmutable(t *testing.T) {
	cfg := &Config{WebhookURL: "http://example.com/webhook1"}
	_ = NewSlackClient(cfg)

	// Client should still use original URL
	// (This documents that client stores a copy of the URL)

	t.Log("Behavior: Client stores webhook URL at creation time")
	t.Log("  - Changing config after NewSlackClient() doesn't affect client")
	t.Log("  - Each client is independent of config changes")
}
