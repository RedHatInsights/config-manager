package util

import (
	"strings"
)

// NormalizeWhitespace removes extra whitespace characters from a string,
// replacing them with a single space, and trims leading and trailing whitespace.
func NormalizeWhitespace(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimSpace(s)
	return s
}
