package validation

import (
	"fmt"
	"strings"

	"github.com/jonesrussell/goforms/internal/core/subscription"
)

// ValidateSubscription validates a subscription request
func ValidateSubscription(sub *subscription.Subscription) error {
	if strings.TrimSpace(sub.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(sub.Email) {
		return fmt.Errorf("invalid email format")
	}

	if strings.TrimSpace(sub.Name) == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}
