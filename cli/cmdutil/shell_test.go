package cmdutil

import "testing"

func TestJoinShellArgs(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"empty", nil, ""},
		{"safe argv passes through", []string{"ls", "-la", "/opt"}, "ls -la /opt"},
		{"semicolon is quoted", []string{"bash", "-c", "echo a; pwd"}, "bash -c 'echo a; pwd'"},
		{"spaces inside arg", []string{"echo", "hello world"}, "echo 'hello world'"},
		{"embedded single quote", []string{"echo", "it's"}, `echo 'it'\''s'`},
		{"dollar var stays inside quotes", []string{"bash", "-c", "echo $HOME"}, "bash -c 'echo $HOME'"},
		{"empty arg becomes ''", []string{"printf", "%s", ""}, "printf %s ''"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := JoinShellArgs(tc.args)
			if got != tc.want {
				t.Errorf("JoinShellArgs(%q) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}
