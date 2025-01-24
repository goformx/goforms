package validation

import (
	"fmt"
	"net/mail"

	"github.com/jonesrussell/goforms/internal/core/contact"
)

// ValidateContact validates a contact form submission
func ValidateContact(sub *contact.Submission) error {
	if sub == nil {
		return fmt.Errorf("submission cannot be nil")
	}

	// Validate email
	if sub.Email == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(sub.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// Validate name
	if sub.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate message
	if sub.Message == "" {
		return fmt.Errorf("message is required")
	}

	return nil
}
