package utils

import (
	"regexp"
	"strings"
)

// CleanHTML sanitizes HTML content by fixing common malformed patterns
// that can cause rendering issues.
func CleanHTML(html string) string {
	if html == "" {
		return html
	}

	// Fix malformed opening tags like <p--> to <p>
	// This pattern matches tags like <p-->, <div-->, etc.
	malformedTagRegex := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)\s*--+>`)
	html = malformedTagRegex.ReplaceAllString(html, "<$1>")

	// Fix malformed self-closing tags like <img--> to <img>
	// Some feeds have broken self-closing tags
	malformedSelfClosingRegex := regexp.MustCompile(`<(img|br|hr|input|meta|link)\s+([^>]*?)--+>`)
	html = malformedSelfClosingRegex.ReplaceAllString(html, "<$1 $2>")

	// Trim any leading/trailing whitespace
	html = strings.TrimSpace(html)

	return html
}
