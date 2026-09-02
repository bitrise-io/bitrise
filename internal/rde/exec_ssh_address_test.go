package rde

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh/agent"
)

func TestParseSSHAddress(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantUser  string
		wantHost  string
		wantPort  int
		wantError bool
	}{
		{name: "bare user@host", addr: "vagrant@host.example", wantUser: "vagrant", wantHost: "host.example", wantPort: 22},
		{name: "ssh command", addr: "ssh vagrant@host.example", wantUser: "vagrant", wantHost: "host.example", wantPort: 22},
		{name: "separate -p", addr: "ssh -p 2222 vagrant@host.example", wantUser: "vagrant", wantHost: "host.example", wantPort: 2222},
		{name: "attached -p", addr: "ssh -p2222 vagrant@host.example", wantUser: "vagrant", wantHost: "host.example", wantPort: 2222},

		// A ":port" suffix used to be dropped: the user@host pattern stopped at
		// the colon and only "-p N" could set a port.
		{name: "user@host:port", addr: "vagrant@host.example:2222", wantUser: "vagrant", wantHost: "host.example", wantPort: 2222},

		// An unanchored "-p\s+(\d+)" search used to match inside the -L
		// argument and report port 80.
		{name: "-p inside another option's argument is not a port", addr: "ssh u@h -L 8080:localhost-p 80", wantUser: "u", wantHost: "h", wantPort: 22},

		// A "\w+@\w+" search used to truncate an IPv6 host at its first colon
		// and silently dial host "2001".
		{name: "bare IPv6 host", addr: "ssh vagrant@2001:db8::1", wantUser: "vagrant", wantHost: "2001:db8::1", wantPort: 22},
		{name: "bracketed IPv6 with port", addr: "vagrant@[2001:db8::1]:2222", wantUser: "vagrant", wantHost: "2001:db8::1", wantPort: 2222},
		{name: "bracketed IPv6 without port", addr: "vagrant@[2001:db8::1]", wantUser: "vagrant", wantHost: "2001:db8::1", wantPort: 22},

		{name: "options before the target are skipped", addr: "ssh -o StrictHostKeyChecking=no -p 2222 ubuntu@h.example", wantUser: "ubuntu", wantHost: "h.example", wantPort: 2222},
		{name: "-p and matching :port agree", addr: "ssh -p 2222 u@h:2222", wantUser: "u", wantHost: "h", wantPort: 2222},

		// An option's argument must not be mistaken for the destination, in
		// either the separate or the attached form.
		{name: "option argument is not the destination", addr: "ssh -o ProxyJump=jump@bastion ubuntu@h.example", wantUser: "ubuntu", wantHost: "h.example", wantPort: 22},
		{name: "attached option argument is not the destination", addr: "ssh -oProxyJump=jump@bastion ubuntu@h.example", wantUser: "ubuntu", wantHost: "h.example", wantPort: 22},

		// The destination is the first operand; the rest is the remote command,
		// so an "@" appearing in it must not retarget the dial.
		{name: "remote command after the destination is ignored", addr: "ssh vagrant@host.example -- echo user@evil", wantUser: "vagrant", wantHost: "host.example", wantPort: 22},
		{name: "trailing operand is not the destination", addr: "ssh -p 2222 vagrant@h extra@operand", wantUser: "vagrant", wantHost: "h", wantPort: 2222},
		{name: "option-looking token after the destination is remote command, not an option", addr: "ssh u@h -p", wantUser: "u", wantHost: "h", wantPort: 22},

		// Clustered short flags: the first letter taking an argument consumes
		// the remainder of the token, or the next one.
		{name: "clustered flags with separate port", addr: "ssh -tp 2222 u@h", wantUser: "u", wantHost: "h", wantPort: 2222},
		{name: "clustered flags with attached port", addr: "ssh -tp2222 u@h", wantUser: "u", wantHost: "h", wantPort: 2222},

		{name: "ssh:// URI", addr: "ssh://vagrant@host.example", wantUser: "vagrant", wantHost: "host.example", wantPort: 22},
		{name: "ssh:// URI with port", addr: "ssh://vagrant@host.example:2222", wantUser: "vagrant", wantHost: "host.example", wantPort: 2222},

		{name: "out of range port", addr: "ssh -p 99999 u@h", wantError: true},
		{name: "zero port", addr: "ssh -p 0 u@h", wantError: true},
		{name: "non-numeric port", addr: "ssh -p abc u@h", wantError: true},
		{name: "-p consuming the destination as its port", addr: "ssh -p u@h", wantError: true},
		{name: "-p missing its argument", addr: "ssh -p", wantError: true},
		{name: "conflicting ports", addr: "ssh -p 2222 u@h:2223", wantError: true},
		{name: "no user component", addr: "host.example:22", wantError: true},
		{name: "empty user", addr: "@host.example", wantError: true},
		{name: "second @ in the destination", addr: "u@h@h2", wantError: true},
		{name: "non-ssh scheme", addr: "http://u@h", wantError: true},
		{name: "bracketed host that is not an IP", addr: "u@[not-an-ip]", wantError: true},
		{name: "option missing its argument", addr: "ssh -o", wantError: true},
		{name: "empty address", addr: "", wantError: true},
		{name: "whitespace only", addr: "   ", wantError: true},
		{name: "only options", addr: "ssh -p 2222", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSSHAddress(tt.addr)
			if tt.wantError {
				if err == nil {
					t.Fatalf("parseSSHAddress(%q) = %+v, want error", tt.addr, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSSHAddress(%q): %v", tt.addr, err)
			}
			if got.User != tt.wantUser || got.Host != tt.wantHost || got.Port != tt.wantPort {
				t.Errorf("parseSSHAddress(%q) = %s@%s:%d, want %s@%s:%d",
					tt.addr, got.User, got.Host, got.Port, tt.wantUser, tt.wantHost, tt.wantPort)
			}
		})
	}
}

func TestSSHTargetForSession(t *testing.T) {
	ready := Session{
		Status:            "running",
		SSHConnectionOpen: true,
		SSHAddress:        "ssh -p 2222 vagrant@host.example",
		SSHPassword:       "hunter2",
	}

	t.Run("ready session yields the dial target", func(t *testing.T) {
		got, err := sshTargetForSession(ready)
		if err != nil {
			t.Fatalf("sshTargetForSession: %v", err)
		}
		want := sshTarget{User: "vagrant", Host: "host.example", Port: 2222, Password: "hunter2"}
		if got != want {
			t.Errorf("target = %+v, want %+v", got, want)
		}
	})

	tests := []struct {
		name     string
		sess     Session
		wantHint string
	}{
		{
			name:     "not running",
			sess:     Session{Status: "stopped", SSHConnectionOpen: true, SSHAddress: ready.SSHAddress, SSHPassword: "p"},
			wantHint: "session is not running",
		},
		{
			name:     "ssh connection not open",
			sess:     Session{Status: "running", SSHAddress: ready.SSHAddress, SSHPassword: "p"},
			wantHint: "session SSH is not ready yet",
		},
		{
			name:     "no address",
			sess:     Session{Status: "running", SSHConnectionOpen: true, SSHPassword: "p"},
			wantHint: "session SSH is not ready yet",
		},
		{
			name:     "no password",
			sess:     Session{Status: "running", SSHConnectionOpen: true, SSHAddress: ready.SSHAddress},
			wantHint: "session SSH is not ready yet",
		},
		{
			name:     "unparsable address",
			sess:     Session{Status: "running", SSHConnectionOpen: true, SSHAddress: "host.example", SSHPassword: "p"},
			wantHint: "parse session ssh address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sshTargetForSession(tt.sess)
			if err == nil {
				t.Fatalf("want error mentioning %q, got nil", tt.wantHint)
			}
			if !strings.Contains(err.Error(), tt.wantHint) {
				t.Errorf("error = %q, want it to mention %q", err, tt.wantHint)
			}
		})
	}
}

// TestSSHAuthMethods asserts how the chain is composed. ssh.AuthMethod values
// are opaque, so the password-first ordering that keeps sshd's MaxAuthTries
// from being exhausted stays a documented invariant rather than an assertion.
func TestSSHAuthMethods(t *testing.T) {
	// Redirected so defaultKeyFilesAuthMethod finds no ~/.ssh keys and the
	// counts don't depend on whoever runs the tests. os.UserHomeDir reads
	// USERPROFILE on Windows and HOME elsewhere, and this file is not build
	// constrained, so both are set.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tests := []struct {
		name     string
		password string
		agent    agent.ExtendedAgent
		want     int
	}{
		{name: "nothing available", want: 0},
		{name: "password only", password: "hunter2", want: 1},
		{name: "agent only", agent: stubAgent{}, want: 1},
		{name: "password and agent", password: "hunter2", agent: stubAgent{}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sshAuthMethods(tt.password, tt.agent); len(got) != tt.want {
				t.Errorf("sshAuthMethods() returned %d methods, want %d", len(got), tt.want)
			}
		})
	}
}

// stubAgent is a non-nil agent.ExtendedAgent for the auth-chain test; only its
// presence matters, so no method is ever called.
type stubAgent struct {
	agent.ExtendedAgent
}
