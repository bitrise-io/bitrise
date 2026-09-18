package cmdutil

import "strings"

// JoinShellArgs joins argv into a single shell-safe command string by
// POSIX-quoting each argument. Without quoting, `bash -c 'echo a; pwd'`
// would arrive at the remote as three space-separated tokens and the `;`
// would be reinterpreted by the remote shell, dropping `echo a`.
func JoinShellArgs(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		q[i] = ShellQuote(a)
	}
	return strings.Join(q, " ")
}

// ShellQuote wraps s in POSIX single-quotes, escaping any embedded single
// quotes with the standard `'\”` sequence. Strings made entirely of
// shell-safe characters are returned as-is to keep readable commands
// readable in logs.
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '_', c == '-', c == '.', c == '/', c == ':', c == '=', c == ',', c == '@', c == '+', c == '%':
		default:
			safe = false
		}
		if !safe {
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
