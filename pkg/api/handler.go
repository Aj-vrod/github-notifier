package api

import (
	"Aj-vrod/github-notifier/internal/storagev0"
	"Aj-vrod/github-notifier/pkg/subscriber"
	"encoding/json"
	"net/http"
)

const (
	successStatus = "subscribed"
)

// SubscribeRequest represents the POST /api/v1/subscribe request body
type SubscribeRequest struct {
	PRURL string `json:"pr_url"`
}

// SubscribeResponse represents the success response for POST /api/v1/subscribe
type SubscribeResponse struct {
	Status string `json:"status"`
	PRURL  string `json:"pr_url"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// writeError writes a consistent error response to the client
func writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// HandleHealth handles GET /health requests
// Returns a simple health check response to verify the service is running
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status: "healthy",
	})
}

// HandleSubscribe handles POST /api/v1/subscribe requests
// Validates the request and returns appropriate success or error responses
func HandleSubscribe(w http.ResponseWriter, r *http.Request, subscriber *subscriber.Subscriber, storage *storagev0.Storage) {
	var req SubscribeRequest

	// Step 1: Request format validation - decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	// Step 2: Request format validation - check if pr_url is provided
	if req.PRURL == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "pr_url is required")
		return
	}

	// Step 3: PR URL format validation
	if err := ValidatePRURL(req.PRURL); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pr_url", err.Error())
		return
	}

	// Step 4: Check if the PR exists on GitHub using the GitHub client
	prInfo, err := ParsePRURL(req.PRURL) // This will extract owner, repo, and PR number
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pr_url", err.Error())
		return
	}
	err = subscriber.Subscribe(r.Context(), prInfo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "github_error", "failed to check PR state on GitHub")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SubscribeResponse{
		Status: successStatus,
		PRURL:  req.PRURL,
	})
}
