package errors

import (
	"errors"
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
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
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
	httpErr = &HTTPError{
		Code:    statusCode,
		Message: GetErrorMessage(err),
		Details: GetErrorContext(err),
	}

	return httpErr
}

// getHTTPStatusCode maps domain error codes to HTTP status codes
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
		ErrCodeBadRequest, ErrCodeFormValidation, ErrCodeFormInvalid, ErrCodeUserInvalid,
		ErrCodeFormSubmission, ErrCodeFormExpired, ErrCodeUserDisabled:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeUserUnauthorized, ErrCodeInvalidToken, ErrCodeAuthentication:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeFormAccessDenied, ErrCodeInsufficientRole:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeFormNotFound, ErrCodeUserNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeAlreadyExists, ErrCodeUserExists:
		return http.StatusConflict
	case ErrCodeServerError, ErrCodeDatabase, ErrCodeConfig:
		return http.StatusInternalServerError
	case ErrCodeStartup, ErrCodeShutdown:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
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
