package sanitization

import (
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Error("NewService() returned nil")
	}
}

func TestService_String(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "XSS script tag",
			input:    "<script>alert('test');</script>",
			expected: ">alert('test');</",
		},
		{
			name:     "normal text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.String(tt.input)
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_Email(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid email",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "email with spaces",
			input:    " test@example.com ",
			expected: "test@example.com",
		},
		{
			name:     "invalid email",
			input:    "invalid-email",
			expected: "invalid-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Email(tt.input)
			if result != tt.expected {
				t.Errorf("Email() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_URL(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid http URL",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "valid https URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "invalid URL",
			input:    "not-a-url",
			expected: "not-a-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.URL(tt.input)
			if result != tt.expected {
				t.Errorf("URL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_HTML(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name:     "script tags",
			input:    "<script>alert('test');</script>",
			expected: "alert('test');",
		},
		{
			name:     "plain text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HTML(tt.input)
			if result != tt.expected {
				t.Errorf("HTML() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_TrimAndSanitize(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with spaces",
			input:    "  Hello, World!  ",
			expected: "Hello, World!",
		},
		{
			name:     "with XSS",
			input:    "  <script>alert('test');</script>  ",
			expected: ">alert('test');</",
		},
		{
			name:     "empty string",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.TrimAndSanitize(tt.input)
			if result != tt.expected {
				t.Errorf("TrimAndSanitize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_TrimAndSanitizeEmail(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with spaces",
			input:    "  test@example.com  ",
			expected: "test@example.com",
		},
		{
			name:     "invalid email with spaces",
			input:    "  invalid-email  ",
			expected: "invalid-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.TrimAndSanitizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("TrimAndSanitizeEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_SanitizeMap(t *testing.T) {
	service := NewService()

	data := map[string]interface{}{
		"name":    "<script>alert('test');</script>",
		"email":   "test@example.com",
		"message": "Hello, World!",
		"nested": map[string]interface{}{
			"field": "<b>Bold</b> text",
		},
	}

	service.SanitizeMap(data)

	if data["name"] != ">alert('test');</" {
		t.Errorf("SanitizeMap() did not sanitize name field correctly")
	}

	if data["email"] != "test@example.com" {
		t.Errorf("SanitizeMap() changed valid email")
	}

	if data["message"] != "Hello, World!" {
		t.Errorf("SanitizeMap() changed plain text")
	}

	nested := data["nested"].(map[string]interface{})
	if nested["field"] != "<b>Bold</b> text" {
		t.Errorf("SanitizeMap() did not sanitize nested field correctly")
	}
}

func TestService_SanitizeFormData(t *testing.T) {
	service := NewService()

	data := map[string]string{
		"name":    "  John Doe  ",
		"email":   "  test@example.com  ",
		"url":     "  http://example.com  ",
		"message": "<script>alert('test');</script>",
	}

	fieldTypes := map[string]string{
		"name":    "string",
		"email":   "email",
		"url":     "url",
		"message": "html",
	}

	result := service.SanitizeFormData(data, fieldTypes)

	if result["name"] != "John Doe" {
		t.Errorf("SanitizeFormData() did not trim and sanitize name correctly")
	}

	if result["email"] != "test@example.com" {
		t.Errorf("SanitizeFormData() did not sanitize email correctly")
	}

	if result["url"] != "http://example.com" {
		t.Errorf("SanitizeFormData() did not sanitize URL correctly")
	}

	if result["message"] != "alert('test');" {
		t.Errorf("SanitizeFormData() did not sanitize HTML correctly")
	}
}

func TestService_SanitizeJSON(t *testing.T) {
	service := NewService()

	data := map[string]interface{}{
		"name":  "<script>alert('test');</script>",
		"email": "test@example.com",
		"tags":  []interface{}{"<b>tag1</b>", "tag2"},
		"nested": map[string]interface{}{
			"field": "<p>content</p>",
		},
	}

	result := service.SanitizeJSON(data).(map[string]interface{})

	if result["name"] != ">alert('test');</" {
		t.Errorf("SanitizeJSON() did not sanitize name field correctly")
	}

	if result["email"] != "test@example.com" {
		t.Errorf("SanitizeJSON() changed valid email")
	}

	tags := result["tags"].([]interface{})
	if tags[0] != "<b>tag1</b>" {
		t.Errorf("SanitizeJSON() did not sanitize array element correctly")
	}

	nested := result["nested"].(map[string]interface{})
	if nested["field"] != "<p>content</p>" {
		t.Errorf("SanitizeJSON() did not sanitize nested field correctly")
	}
}

func TestService_SanitizeWithOptions(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		opts     SanitizeOptions
		expected string
	}{
		{
			name:  "trim whitespace",
			input: "  Hello, World!  ",
			opts: SanitizeOptions{
				TrimWhitespace: true,
			},
			expected: "Hello, World!",
		},
		{
			name:  "remove HTML",
			input: "<p>Hello <b>World</b></p>",
			opts: SanitizeOptions{
				RemoveHTML: true,
			},
			expected: "Hello World",
		},
		{
			name:  "max length",
			input: "This is a very long string that should be truncated",
			opts: SanitizeOptions{
				MaxLength: 20,
			},
			expected: "This is a very long ",
		},
		{
			name:  "all options",
			input: "  <p>Hello <b>World</b></p>  ",
			opts: SanitizeOptions{
				TrimWhitespace: true,
				RemoveHTML:     true,
				MaxLength:      10,
			},
			expected: "Hello Worl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.SanitizeWithOptions(tt.input, tt.opts)
			if result != tt.expected {
				t.Errorf("SanitizeWithOptions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_IsValidEmail(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid email",
			input:    "test@example.com",
			expected: true,
		},
		{
			name:     "invalid email",
			input:    "invalid-email",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsValidEmail(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_IsValidURL(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid http URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "valid https URL",
			input:    "https://example.com",
			expected: true,
		},
		{
			name:     "invalid URL",
			input:    "not-a-url",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsValidURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}
