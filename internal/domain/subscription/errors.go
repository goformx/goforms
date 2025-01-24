package subscription

import "errors"

var (
	// ErrSubscriptionNotFound indicates that a subscription was not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrEmailAlreadyExists indicates that a subscription with the given email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidStatus indicates that the provided status is invalid
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidSubscription indicates that the subscription is invalid
	ErrInvalidSubscription = errors.New("invalid subscription")
	// ErrEmailRequired indicates that email is required
	ErrEmailRequired = errors.New("email is required")
	// ErrNameRequired indicates that name is required
	ErrNameRequired = errors.New("name is required")
	// ErrInvalidEmail indicates that email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
)
