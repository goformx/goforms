package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jonesrussell/goforms/internal/core/contact"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateContact validates a contact submission
func ValidateContact(submission *contact.Submission) error {
	if submission == nil {
		return fmt.Errorf("submission cannot be nil")
	}

	if strings.TrimSpace(submission.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(submission.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(submission.Email) {
		return fmt.Errorf("invalid email format")
	}

	if strings.TrimSpace(submission.Message) == "" {
		return fmt.Errorf("message is required")
	}

	return nil
}
