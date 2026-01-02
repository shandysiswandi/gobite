package stacktrace

import "strings"

// InternalPaths returns internal package stack frames from a raw stack trace.
func InternalPaths(stack []byte) []string {
	lines := strings.Split(string(stack), "\n")
	paths := make([]string, 0, len(lines))
	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i+1])
		if strings.Contains(line, "/internal/") && strings.Contains(line, ".go") {
			if idx := strings.Index(line, ".go:"); idx != -1 {
				end := strings.Index(line[idx:], " ")
				if end == -1 {
					end = len(line)
				} else {
					end += idx
				}
				shortPath := line[:end]
				internalIdx := strings.Index(shortPath, "/internal/")
				if internalIdx != -1 {
					shortPath = shortPath[internalIdx+1:]
					paths = append(paths, shortPath)
				}
			}
		}
	}
	return paths
}
