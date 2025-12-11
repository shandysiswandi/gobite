package pkgstrcase

import "strings"

// ToLowerSnake converts a string to snake_case.
func ToLowerSnake(s string) string {
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}

	return strings.ToLower(string(result))
}
