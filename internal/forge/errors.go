// Package forge provides a client for the Laravel Forge API.
package forge

import "fmt"

// APIError represents a non-2xx response from the Forge API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("forge API error %d: %s", e.StatusCode, e.Message)
}

// AuthenticationError is returned when the API responds with 401 Unauthorized.
type AuthenticationError struct{ APIError }

// NotFoundError is returned when the API responds with 404 Not Found.
type NotFoundError struct{ APIError }

// RateLimitError is returned when the API responds with 429 Too Many Requests.
type RateLimitError struct{ APIError }

// ValidationError is returned when the API responds with 422 Unprocessable Entity.
// Details contains per-field validation messages keyed by field name.
type ValidationError struct {
	APIError
	Details map[string][]string
}
