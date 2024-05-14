package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware Handles generating request id and capturing status for logging
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract or generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID if not present in headers
			timestamp := time.Now().UnixNano()

			// Convert timestamp to a string
			requestID = fmt.Sprintf("%d", timestamp)
		}

		// Add request ID to response headers
		r.Header.Set("X-Request-ID", requestID)

		// Create a custom ResponseWriter to capture the response status code
		rw := responseWriter{w, http.StatusOK}

		// Call the next handler function with the custom ResponseWriter
		next(&rw, r)

		// Log the request details
		duration := time.Since(start)
		log.Printf("[%s] %s %s %s %d %s\n", requestID, r.Method, r.URL.Path, r.RemoteAddr, rw.status, duration)
	}
}

// Define a custom ResponseWriter to capture the response status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// Override the WriteHeader method to capture the response status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
