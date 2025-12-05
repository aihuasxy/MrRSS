package utils

import "testing"

func TestNormalizeURLForComparison(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without query params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "URL with tracking params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Edge&utm_medium=rss&fc=20251205031104",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "URL with different tracking params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/?fc=20251206000000&ff=20251206000001&v=2.18.0.post22",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "HTTP URL with params",
			input:    "http://example.com/article/123?tracking=abc",
			expected: "http://example.com/article/123",
		},
		{
			name:     "Invalid URL returns original",
			input:    "not-a-valid-url",
			expected: "not-a-valid-url",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURLForComparison(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURLForComparison(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLsMatch(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "Identical URLs",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: true,
		},
		{
			name:     "Same base URL with different tracking params",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Edge&fc=20251205031104",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Chrome&fc=20251206000000",
			expected: true,
		},
		{
			name:     "One with params one without",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?fc=20251205031104",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: true,
		},
		{
			name:     "Different article IDs",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345846/",
			expected: false,
		},
		{
			name:     "Different hosts",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://example.com/41345845/",
			expected: false,
		},
		{
			name:     "Empty strings",
			url1:     "",
			url2:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URLsMatch(tt.url1, tt.url2)
			if result != tt.expected {
				t.Errorf("URLsMatch(%q, %q) = %v, want %v", tt.url1, tt.url2, result, tt.expected)
			}
		})
	}
}
