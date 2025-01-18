package validation

import (
	"fmt"
	"strings"

	"github.com/jonesrussell/goforms/internal/models"
)

// ValidateSubscription validates a subscription request
func ValidateSubscription(sub *models.Subscription) error {
	if strings.TrimSpace(sub.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(sub.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
