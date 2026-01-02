package strcase

import (
	"strings"
	"unicode"
)

// ToLowerSnake converts a string to snake_case (initialism-safe).
func ToLowerSnake(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	runes := []rune(s)

	for i := range runes {
		r := runes[i]

		// Add underscore at word boundaries:
		// 1) lower/digit -> upper  (e.g., userID -> user_ID)
		// 2) acronym -> word       (e.g., HTTPServer -> HTTP_Server)
		if i > 0 {
			prev := runes[i-1]
			var next rune
			if i+1 < len(runes) {
				next = runes[i+1]
			}

			if unicode.IsUpper(r) {
				// case 1: prev is lower or digit
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteRune('_')
				} else if unicode.IsUpper(prev) && next != 0 && unicode.IsLower(next) {
					// case 2: boundary between acronym and word
					b.WriteRune('_')
				}
			}
		}

		b.WriteRune(unicode.ToLower(r))
	}

	return b.String()
}
