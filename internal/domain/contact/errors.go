package contact

import "errors"

var (
	// ErrSubmissionNotFound is returned when a contact submission is not found
	ErrSubmissionNotFound = errors.New("contact submission not found")
)
