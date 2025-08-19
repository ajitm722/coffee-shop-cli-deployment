package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee/internal/server"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// TestHealthEndpoints verifies the health and readiness endpoints of the server.
// It ensures that the server responds with HTTP 200 for both endpoints, indicating proper functionality.
func TestHealthEndpoints(t *testing.T) {
	// Create a new server instance with default logger.
	s := server.NewServer(
		server.WithLogger(slog.Default()), // Use the default logger for structured logging.
		server.WithGinMode(gin.TestMode),  // Set Gin mode to TestMode for testing purposes.
	)

	eng := s.Engine() // Get the Gin engine from the server.

	// Test the /v1/healthz endpoint.
	req := httptest.NewRequest(http.MethodGet, "/v1/healthz", nil) // Create a GET request for the health check.
	w := httptest.NewRecorder()                                    // Record the response.
	eng.ServeHTTP(w, req)                                          // Serve the request using the Gin engine.
	if w.Code != http.StatusOK {                                   // Check if the response status code is 200.
		t.Fatalf("healthz expected 200, got %d", w.Code)
	}

	// Test the /v1/readyz endpoint.
	req2 := httptest.NewRequest(http.MethodGet, "/v1/readyz", nil) // Create another GET request for readiness check.
	w2 := httptest.NewRecorder()                                   // Record the response.
	eng.ServeHTTP(w2, req2)                                        // Serve the request using the Gin engine.
	if w2.Code != http.StatusOK {                                  // Check if the response status code is 200.
		t.Fatalf("readyz expected 200, got %d", w2.Code)
	}
}
