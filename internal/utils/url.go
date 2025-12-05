package utils

import "net/url"

// NormalizeURLForComparison returns a normalized URL for comparison purposes.
// It strips query parameters that often change between feed fetches (like tracking params).
// This helps match articles even when feeds use dynamic URL parameters.
func NormalizeURLForComparison(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	// If no scheme, return original (likely invalid URL)
	if parsed.Scheme == "" {
		return rawURL
	}
	// Return just scheme + host + path (without query parameters)
	return parsed.Scheme + "://" + parsed.Host + parsed.Path
}

// URLsMatch checks if two URLs refer to the same article by comparing their normalized forms.
// It first tries exact match, then falls back to normalized comparison.
func URLsMatch(url1, url2 string) bool {
	// Try exact match first
	if url1 == url2 {
		return true
	}
	// Fall back to normalized comparison (ignoring query parameters)
	return NormalizeURLForComparison(url1) == NormalizeURLForComparison(url2)
}
