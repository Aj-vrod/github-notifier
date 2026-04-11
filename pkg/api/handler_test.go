package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestHandleSubscribe tests the POST /api/v1/subscribe endpoint
func TestHandleSubscribe(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   SuccessResponse
	}{
		{
			name:           "valid request",
			requestBody:    `{"pr_url": "https://github.com/owner/repo/pull/123"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   SuccessResponse{Success: "ok"},
		},
		{
			name:           "valid request with different URL",
			requestBody:    `{"pr_url": "https://github.com/facebook/react/pull/456"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   SuccessResponse{Success: "ok"},
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"pr_url": invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   SuccessResponse{},
		},
		{
			name:           "empty body",
			requestBody:    ``,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   SuccessResponse{},
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
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Only check response body for successful requests
			if tt.expectedStatus == http.StatusOK {
				// Check Content-Type header
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}

				// Check response body
				var response SuccessResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if response.Success != tt.expectedBody.Success {
					t.Errorf("expected success '%s', got '%s'", tt.expectedBody.Success, response.Success)
				}
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
			expectedStatus: http.StatusOK,
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
