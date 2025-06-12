package shared

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
