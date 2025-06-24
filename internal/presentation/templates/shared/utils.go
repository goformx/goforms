package shared

import (
	"strings"

	"github.com/goformx/goforms/internal/domain/form/model"
)

// GetMessageIcon returns the appropriate Bootstrap icon class for the message type
func GetMessageIcon(msgType string) string {
	switch msgType {
	case "error":
		return "exclamation-circle"
	case "success":
		return "check-circle"
	case "warning":
		return "exclamation-triangle"
	default:
		return "info-circle"
	}
}

// GetCorsOriginsString extracts origins from the CORS JSON and returns them as a comma-separated string
func GetCorsOriginsString(corsOrigins model.JSON) string {
	if corsOrigins == nil {
		return ""
	}

	if originsArr, ok := corsOrigins["origins"].([]any); ok {
		var origins []string
		for _, origin := range originsArr {
			if originStr, originOk := origin.(string); originOk {
				origins = append(origins, originStr)
			}
		}
		return strings.Join(origins, ",")
	}

	return ""
}
