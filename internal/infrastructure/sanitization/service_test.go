package sanitization_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

func TestNewService(t *testing.T) {
	service := sanitization.NewService()
	if service == nil {
		t.Error("NewService() returned nil")
	}
}

func TestService_String(t *testing.T) {
	service := sanitization.NewService()

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
	service := sanitization.NewService()

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
	service := sanitization.NewService()

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
	service := sanitization.NewService()

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
	service := sanitization.NewService()

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
			expected: "alert('test');",
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
	service := sanitization.NewService()

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
	service := sanitization.NewService()

	data := map[string]any{
		"name":  "  John Doe  ",
		"email": "  test@example.com  ",
		"nested": map[string]any{
			"title": "  <b>Test Title</b>  ",
		},
	}

	service.SanitizeMap(data)

	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "test@example.com", data["email"])
	nested, ok := data["nested"].(map[string]any)
	require.True(t, ok, "expected nested to be map[string]any")
	assert.Equal(t, "Test Title", nested["title"])
}

func TestService_SanitizeFormData(t *testing.T) {
	service := sanitization.NewService()

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
	service := sanitization.NewService()

	data := map[string]any{
		"name":  "  John Doe  ",
		"email": "  test@example.com  ",
		"tags":  []any{"<b>tag1</b>", "tag2"},
		"nested": map[string]any{
			"title": "  <b>Test Title</b>  ",
		},
	}

	result, ok := service.SanitizeJSON(data).(map[string]any)
	require.True(t, ok, "expected result to be map[string]any")

	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, "test@example.com", result["email"])
	tags, ok := result["tags"].([]any)
	require.True(t, ok, "expected tags to be []any")
	assert.Equal(t, "tag1", tags[0])
	assert.Equal(t, "tag2", tags[1])

	nested, ok := result["nested"].(map[string]any)
	require.True(t, ok, "expected nested to be map[string]any")
	assert.Equal(t, "Test Title", nested["title"])
}

func TestService_SanitizeWithOptions(t *testing.T) {
	service := sanitization.NewService()

	tests := []struct {
		name     string
		input    string
		opts     sanitization.SanitizeOptions
		expected string
	}{
		{
			name:  "trim whitespace",
			input: "  Hello, World!  ",
			opts: sanitization.SanitizeOptions{
				TrimWhitespace: true,
			},
			expected: "Hello, World!",
		},
		{
			name:  "remove HTML",
			input: "<p>Hello <b>World</b></p>",
			opts: sanitization.SanitizeOptions{
				RemoveHTML: true,
			},
			expected: "Hello World",
		},
		{
			name:  "max length",
			input: "This is a very long string that should be truncated",
			opts: sanitization.SanitizeOptions{
				MaxLength: 20,
			},
			expected: "This is a very long ",
		},
		{
			name:  "all options",
			input: "  <p>Hello <b>World</b></p>  ",
			opts: sanitization.SanitizeOptions{
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
	service := sanitization.NewService()

	tests := []struct {
		input    string
		expected bool
	}{
		{"test@example.com", true},
		{"invalid-email", false},
		{"", false},
	}

	for _, tt := range tests {
		result := sanitization.IsValidEmail(service, tt.input)
		if result != tt.expected {
			t.Errorf("IsValidEmail(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestService_IsValidURL(t *testing.T) {
	service := sanitization.NewService()

	tests := []struct {
		input    string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"ftp://example.com", false},
		{"invalid-url", false},
		{"", false},
	}

	for _, tt := range tests {
		result := sanitization.IsValidURL(service, tt.input)
		if result != tt.expected {
			t.Errorf("IsValidURL(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestService_SanitizeSlice(t *testing.T) {
	service := sanitization.NewService()

	data := []any{
		"  item1  ",
		"  <b>item2</b>  ",
		"  item3  ",
	}

	service.SanitizeSlice(data)

	assert.Equal(t, "item1", data[0])
	assert.Equal(t, "item2", data[1])
	assert.Equal(t, "item3", data[2])
}

func TestService_SanitizeForLogging(t *testing.T) {
	service := sanitization.NewService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline injection",
			input:    "normal message\nmalicious log entry",
			expected: "normal message malicious log entry",
		},
		{
			name:     "carriage return injection",
			input:    "normal message\rmalicious log entry",
			expected: "normal message malicious log entry",
		},
		{
			name:     "null byte injection",
			input:    "normal message\x00malicious log entry",
			expected: "normal messagemalicious log entry",
		},
		{
			name:     "HTML injection",
			input:    "normal message<script>alert('xss')</script>",
			expected: "normal message&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "mixed injection",
			input:    "normal\nmessage<script>alert('xss')</script>\r\n",
			expected: "normal message&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "HTML entities",
			input:    "message with < and > and &",
			expected: "message with &lt; and &gt; and &amp;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n\r   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.SanitizeForLogging(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
