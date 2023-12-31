package fctl

import "strings"

func ConcatPath(paths ...string) string {
	sb := strings.Builder{}
	for i, p := range paths {
		trimmedPath := strings.TrimSuffix(p, "/")
		sb.WriteString(trimmedPath)
		if i < len(paths)-1 {
			sb.WriteRune('/')
		}
	}
	return sb.String()
}
