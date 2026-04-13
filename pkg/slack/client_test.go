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
