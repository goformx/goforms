package errors

import (
	"net/http"
)

// HTTPError represents an HTTP error response
type HTTPError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// TranslateToHTTP translates a domain error to an HTTP error
func TranslateToHTTP(err error) *HTTPError {
	if err == nil {
		return nil
	}

	// If it's already an HTTPError, return it
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr
	}

	// Get the domain error code
	code := GetErrorCode(err)
	if code == "" {
		// If it's not a domain error, return a generic server error
		return &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		}
	}

	// Translate the domain error code to an HTTP status code
	statusCode := getHTTPStatusCode(code)

	// Create the HTTP error
	httpErr := &HTTPError{
		Code:    statusCode,
		Message: GetErrorMessage(err),
		Details: GetErrorContext(err),
	}

	return httpErr
}

// getHTTPStatusCode maps domain error codes to HTTP status codes
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeAuthentication:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInsufficientRole:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeAlreadyExists:
		return http.StatusConflict
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeServerError, ErrCodeDatabase, ErrCodeTimeout:
		return http.StatusInternalServerError
	case ErrCodeStartup, ErrCodeShutdown, ErrCodeConfig:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		Code:    statusCode,
		Message: message,
	}
}

// WithDetails adds details to the HTTP error
func (e *HTTPError) WithDetails(details map[string]any) *HTTPError {
	e.Details = details
	return e
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return e.Message
}
