// Package server contains the server-related functionality, including route registration.
package server

import (
	"database/sql"
	"net/http" // Provides HTTP status codes and related utilities.

	"github.com/gin-gonic/gin" // Gin is a web framework for building HTTP APIs.
	"github.com/lib/pq"        // pq is a PostgreSQL driver for Go, used for database interactions.
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

// Register menu routes
func registerMenuRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/menu", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, price_cents FROM menu")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var items []gin.H
		for rows.Next() {
			var id, price int
			var name string
			if err := rows.Scan(&id, &name, &price); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, gin.H{
				"id": id, "name": name, "price_cents": price,
			})
		}
		c.JSON(http.StatusOK, items)
	})
}

// Register order routes
func registerOrderRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.POST("/orders", func(c *gin.Context) {
		var payload struct {
			Customer string   `json:"customer"`
			Items    []string `json:"items"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		_, err := db.Exec(
			"INSERT INTO orders (customer_name, items) VALUES ($1, $2)",
			payload.Customer, pq.Array(payload.Items),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "order received",
			"customer": payload.Customer,
		})
	})
}
