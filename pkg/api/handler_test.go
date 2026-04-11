package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandleHealth tests the GET /health endpoint
func TestHandleHealth(t *testing.T) {
	// Create a test HTTP request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler
	HandleHealth(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check response body
	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

// TestHandleSubscribe_Success tests successful POST /api/v1/subscribe requests
func TestHandleSubscribe_Success(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectedURL string
	}{
		{
			name:        "valid request",
			requestBody: `{"pr_url": "https://github.com/owner/repo/pull/123"}`,
			expectedURL: "https://github.com/owner/repo/pull/123",
		},
		{
			name:        "valid request with different URL",
			requestBody: `{"pr_url": "https://github.com/facebook/react/pull/456"}`,
			expectedURL: "https://github.com/facebook/react/pull/456",
		},
		{
			name:        "valid request with repo containing dots",
			requestBody: `{"pr_url": "https://github.com/user/repo.name/pull/999"}`,
			expectedURL: "https://github.com/user/repo.name/pull/999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/subscribe",
				bytes.NewBufferString(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call the handler
			HandleSubscribe(w, req)

			// Check status code
			if w.Code != http.StatusCreated {
				t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
			}

			// Check Content-Type header
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Check response body
			var response SubscribeResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if response.Status != successStatus {
				t.Errorf("expected status '%s', got '%s'", successStatus, response.Status)
			}

			if response.PRURL != tt.expectedURL {
				t.Errorf("expected pr_url '%s', got '%s'", tt.expectedURL, response.PRURL)
			}
		})
	}
}

// TestHandleSubscribe_InvalidJSON tests requests with invalid JSON
func TestHandleSubscribe_InvalidJSON(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "malformed JSON",
			requestBody: `{"pr_url": invalid}`,
		},
		{
			name:        "incomplete JSON",
			requestBody: `{"pr_url": "https://github.com/owner/repo/pull/123"`,
		},
		{
			name:        "empty body",
			requestBody: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/subscribe",
				bytes.NewBufferString(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleSubscribe(w, req)

			// Check status code
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}

			// Check error response
			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode error response body: %v", err)
			}

			if response.Error != "invalid_request" {
				t.Errorf("expected error code 'invalid_request', got '%s'", response.Error)
			}

			if response.Message != "request body must be valid JSON" {
				t.Errorf("expected message about invalid JSON, got '%s'", response.Message)
			}
		})
	}
}

// TestHandleSubscribe_MissingPRURL tests requests with missing pr_url field
func TestHandleSubscribe_MissingPRURL(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "empty pr_url",
			requestBody: `{"pr_url": ""}`,
		},
		{
			name:        "missing pr_url field",
			requestBody: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/subscribe",
				bytes.NewBufferString(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleSubscribe(w, req)

			// Check status code
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}

			// Check error response
			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode error response body: %v", err)
			}

			if response.Error != "invalid_request" {
				t.Errorf("expected error code 'invalid_request', got '%s'", response.Error)
			}

			if response.Message != "pr_url is required" {
				t.Errorf("expected message 'pr_url is required', got '%s'", response.Message)
			}
		})
	}
}

// TestHandleSubscribe_InvalidPRURL tests requests with invalid PR URL formats
func TestHandleSubscribe_InvalidPRURL(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedErrMsg string
	}{
		{
			name:           "http protocol",
			requestBody:    `{"pr_url": "http://github.com/owner/repo/pull/123"}`,
			expectedErrMsg: "pr_url must use https protocol",
		},
		{
			name:           "wrong domain",
			requestBody:    `{"pr_url": "https://gitlab.com/owner/repo/pull/123"}`,
			expectedErrMsg: "pr_url must be from github.com domain",
		},
		{
			name:           "missing repo",
			requestBody:    `{"pr_url": "https://github.com/owner/pull/123"}`,
			expectedErrMsg: "pr_url must match format: https://github.com/{owner}/{repo}/pull/{number}",
		},
		{
			name:           "issues instead of pull",
			requestBody:    `{"pr_url": "https://github.com/owner/repo/issues/123"}`,
			expectedErrMsg: "pr_url must match format: https://github.com/{owner}/{repo}/pull/{number}",
		},
		{
			name:           "invalid owner format",
			requestBody:    `{"pr_url": "https://github.com/owner_name/repo/pull/123"}`,
			expectedErrMsg: "owner must contain only alphanumeric characters and hyphens",
		},
		{
			name:           "owner too long",
			requestBody:    `{"pr_url": "https://github.com/` + strings.Repeat("a", 40) + `/repo/pull/123"}`,
			expectedErrMsg: "owner must not exceed 39 characters",
		},
		{
			name:           "repo too long",
			requestBody:    `{"pr_url": "https://github.com/owner/` + strings.Repeat("a", 101) + `/pull/123"}`,
			expectedErrMsg: "repo must not exceed 100 characters",
		},
		{
			name:           "invalid PR number",
			requestBody:    `{"pr_url": "https://github.com/owner/repo/pull/abc"}`,
			expectedErrMsg: "pull request number must be a positive integer",
		},
		{
			name:           "zero PR number",
			requestBody:    `{"pr_url": "https://github.com/owner/repo/pull/0"}`,
			expectedErrMsg: "pull request number must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/subscribe",
				bytes.NewBufferString(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleSubscribe(w, req)

			// Check status code
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}

			// Check error response
			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode error response body: %v", err)
			}

			if response.Error != "invalid_pr_url" {
				t.Errorf("expected error code 'invalid_pr_url', got '%s'", response.Error)
			}

			if response.Message != tt.expectedErrMsg {
				t.Errorf("expected message '%s', got '%s'", tt.expectedErrMsg, response.Message)
			}
		})
	}
}

// TestServerRoutes tests that the server routes are properly configured
func TestServerRoutes(t *testing.T) {
	server := NewServer(8001)
	router := server.GetRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET /health",
			method:         http.MethodGet,
			path:           "/health",
			body:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /api/v1/subscribe",
			method:         http.MethodPost,
			path:           "/api/v1/subscribe",
			body:           `{"pr_url": "https://github.com/owner/repo/pull/123"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST /health should fail (wrong method)",
			method:         http.MethodPost,
			path:           "/health",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET /api/v1/subscribe should fail (wrong method)",
			method:         http.MethodGet,
			path:           "/api/v1/subscribe",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
