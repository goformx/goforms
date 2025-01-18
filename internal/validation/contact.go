package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jonesrussell/goforms/internal/models"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateContact validates a contact submission
func ValidateContact(contact *models.ContactSubmission) error {
	if strings.TrimSpace(contact.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(contact.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(contact.Email) {
		return fmt.Errorf("invalid email format")
	}

	if strings.TrimSpace(contact.Message) == "" {
		return fmt.Errorf("message is required")
	}

	return nil
}
