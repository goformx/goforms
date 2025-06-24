package sanitization

// Constants for field validation and sanitization
const (
	// MaxStringLength represents the maximum length for string fields
	MaxStringLength = 1000
	// MaxPathLength represents the maximum length for path fields
	MaxPathLength = 500
	// UUIDLength represents the standard UUID length
	UUIDLength = 36
	// UUIDParts represents the number of parts in a UUID
	UUIDParts = 5
	// UUIDMinMaskLen represents the minimum length for UUID masking
	UUIDMinMaskLen = 8
	// UUIDMaskPrefixLen represents the prefix length for UUID masking
	UUIDMaskPrefixLen = 4
	// UUIDMaskSuffixLen represents the suffix length for UUID masking
	UUIDMaskSuffixLen = 4
)
