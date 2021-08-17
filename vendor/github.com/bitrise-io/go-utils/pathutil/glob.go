package pathutil

// EscapeGlobPath escapes a partial path, determined at runtime, used as a parameter for filepath.Glob
func EscapeGlobPath(path string) string {
	var escaped string
	for _, ch := range path {
		if ch == '[' || ch == ']' || ch == '-' || ch == '*' || ch == '?' || ch == '\\' {
			escaped += "\\"
		}
		escaped += string(ch)
	}
	return escaped
}
