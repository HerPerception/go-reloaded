package processor

import "testing"

func TestProcessText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Hexadecimal Conversion",
			input:    "1E (hex) files were added",
			expected: "30 files were added",
		},
		{
			name:     "Binary Conversion",
			input:    "It has been 10 (bin) years",
			expected: "It has been 2 years",
		},
		{
			name:     "Punctuation Spacing Case",
			input:    "I was sitting over there ,and then BAMM !!",
			expected: "I was sitting over there, and then BAMM!!",
		},
		{
			name:     "Article Correction Case",
			input:    "There it was. A amazing rock!",
			expected: "There it was. An amazing rock!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessText(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
