package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// AppError represents a structured application error
type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewNotFound creates a NOT_FOUND error (404)
func NewNotFound(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: resource + " not found",
		Details: nil,
	}
}

// NewBadRequest creates a BAD_REQUEST error (400)
func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Details: nil,
	}
}

// NewConflict creates a CONFLICT error (409)
func NewConflict(message string) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Details: nil,
	}
}

// NewInternalError creates an INTERNAL_ERROR (500)
// Note: This returns a generic message to avoid exposing internal details
func NewInternalError() *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: "An internal error occurred. Please try again later.",
		Details: nil,
	}
}

// NewUnauthorized creates an UNAUTHORIZED error (401)
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Details: nil,
	}
}

// getStatusCode maps AppError codes to HTTP status codes
func getStatusCode(err *AppError) int {
	switch err.Code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "CONFLICT":
		return http.StatusConflict
	case "INTERNAL_ERROR":
		return http.StatusInternalServerError
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "23505")
}

// ErrorHandler intercepts errors and converts them to JSON responses
// This should be registered as the LAST middleware in the chain
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This c is a gin.Context object
		// It's automatically passed by Gin when the middleware is called

		// Process the request
		c.Next()

		// Check if any errors occurred during request processing
		if len(c.Errors) > 0 {
			// Get the last error (most recent)
			err := c.Errors.Last().Err

			// Log the error for debugging
			log.Printf("[ERROR] %s %s - %v (type: %T)", c.Request.Method, c.Request.URL.Path, err, err)

			// Convert error to appropriate HTTP response
			handleError(c, err)
		}
	}
}

// handleError converts any error into an appropriate HTTP JSON response
func handleError(c *gin.Context, err error) {
	// Prevent multiple responses
	if c.Writer.Written() {
		return
	}

	// 1. Check if it's already an AppError (Checks if err is an *AppError)
	if appErr, ok := err.(*AppError); ok {
		c.JSON(getStatusCode(appErr), gin.H{"error": appErr})
		return
	}

	// 2. Check for pgx.ErrNoRows (resource not found)
	if errors.Is(err, pgx.ErrNoRows) {
		notFoundErr := NewNotFound("resource")
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundErr})
		return
	}

	// 3. Check for unique constraint violations
	if isUniqueViolation(err) {
		conflictErr := NewConflict("resource already exists")
		c.JSON(http.StatusConflict, gin.H{"error": conflictErr})
		return
	}

	// 4. Check for validation errors (from Gin binding)
	errMsg := err.Error()
	if strings.Contains(errMsg, "binding") ||
		strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "invalid") {
		badReqErr := NewBadRequest(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": badReqErr})
		return
	}

	// 5. Handle unknown errors - don't expose internal details
	internalErr := NewInternalError()
	c.JSON(http.StatusInternalServerError, gin.H{"error": internalErr})
}
