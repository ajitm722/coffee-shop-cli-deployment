// Package server contains the server-related functionality, including route registration.
package server

import (
	"net/http" // Provides HTTP status codes and related utilities.

	"github.com/gin-gonic/gin" // Gin is a web framework for building HTTP APIs.
)

// registerHealthRoutes registers health check routes to the provided router group.
// These routes are used to monitor the application's health and readiness.
func registerHealthRoutes(rg *gin.RouterGroup) {
	// Define a health check endpoint at /healthz.
	// Responds with HTTP 200 and a JSON object indicating the service is healthy.
	rg.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Define a readiness check endpoint at /readyz.
	// Responds with HTTP 200 and a JSON object indicating the service is ready.
	rg.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
