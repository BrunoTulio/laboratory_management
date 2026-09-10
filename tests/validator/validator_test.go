package validator_test

import (
	"testing"
)

// TestInputValidation tests basic input validation functions
func TestInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		validator func(string) bool
		expected  bool
	}{
		{
			name:  "Valid email",
			input: "test@example.com",
			validator: func(s string) bool {
				// Simple email validation
				return len(s) > 5 && contains(s, "@") && contains(s, ".")
			},
			expected: true,
		},
		{
			name:  "Invalid email - no @",
			input: "testexample.com",
			validator: func(s string) bool {
				return len(s) > 5 && contains(s, "@") && contains(s, ".")
			},
			expected: false,
		},
		{
			name:  "Valid phone",
			input: "(11) 99999-9999",
			validator: func(s string) bool {
				return len(s) >= 10
			},
			expected: true,
		},
		{
			name:  "Invalid phone - too short",
			input: "123",
			validator: func(s string) bool {
				return len(s) >= 10
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validator(tt.input)
			if result != tt.expected {
				t.Errorf("Validator returned %v, want %v", result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}