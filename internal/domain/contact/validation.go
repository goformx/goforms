package contact

import (
	"errors"
	"fmt"
	"net/mail"
)

// ValidateSubmission validates a contact form submission
func ValidateSubmission(sub *Submission) error {
	if sub == nil {
		return errors.New("submission cannot be nil")
	}

	// Validate email
	if sub.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(sub.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// Validate name
	if sub.Name == "" {
		return errors.New("name is required")
	}

	// Validate message
	if sub.Message == "" {
		return errors.New("message is required")
	}

	return nil
}
