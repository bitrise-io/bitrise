package rde

import "testing"

func TestStripInteractiveBashStartupNoise(t *testing.T) {
	const noise1 = "bash: cannot set terminal process group (-1): Inappropriate ioctl for device"
	const noise2 = "bash: no job control in this shell"

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty stays empty",
			in:   "",
			want: "",
		},
		{
			name: "both noise lines, no real output",
			in:   noise1 + "\n" + noise2 + "\n",
			want: "",
		},
		{
			name: "noise then real stderr is preserved verbatim",
			in:   noise1 + "\n" + noise2 + "\nreal error: boom\n",
			want: "real error: boom\n",
		},
		{
			name: "noise in either order is stripped",
			in:   noise2 + "\n" + noise1 + "\nwarning\n",
			want: "warning\n",
		},
		{
			name: "real output untouched when no noise present",
			in:   "just an error\n",
			want: "just an error\n",
		},
		{
			name: "identical line after real output is not stripped (only leading run)",
			in:   "step 1\n" + noise2 + "\n",
			want: "step 1\n" + noise2 + "\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripInteractiveBashStartupNoise(tc.in); got != tc.want {
				t.Errorf("stripInteractiveBashStartupNoise(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildLoginShellCmd(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain command is wrapped verbatim",
			in:   "echo hello",
			want: `bash -i -l -c 'echo hello'`,
		},
		{
			// The command layer's --shell mode joins tokens verbatim, so a
			// command carrying single quotes reaches here unescaped. Each '
			// must become '\'' or it would close the wrapper's quoting early
			// and the remote bash would mis-parse the line.
			name: "single quotes are escaped so the wrapper stays balanced",
			in:   "echo 'hi'",
			want: `bash -i -l -c 'echo '\''hi'\'''`,
		},
		{
			name: "shell metacharacters plus a quoted arg survive intact",
			in:   `cd repo && grep -rn 'TODO' .`,
			want: `bash -i -l -c 'cd repo && grep -rn '\''TODO'\'' .'`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildLoginShellCmd(tc.in); got != tc.want {
				t.Errorf("buildLoginShellCmd(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
