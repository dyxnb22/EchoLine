package middleware

import (
	"regexp"
	"strings"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// StripHTML removes HTML tags from s and trims surrounding whitespace.
func StripHTML(s string) string {
	cleaned := htmlTagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(cleaned)
}

// SanitizeBody strips HTML from a message body before persistence.
func SanitizeBody(body string) string {
	return StripHTML(body)
}
