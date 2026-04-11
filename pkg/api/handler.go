package api

import (
	"encoding/json"
	"net/http"
)

// SubscribeRequest represents the POST /api/v1/subscribe request body
type SubscribeRequest struct {
	PRURL string `json:"pr_url"`
}

// SuccessResponse represents the temporary success response for MS1
type SuccessResponse struct {
	Success string `json:"success"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
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
func HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// For MS1: just return success without any validation or processing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: "ok",
	})
}
