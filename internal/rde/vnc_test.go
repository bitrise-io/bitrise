package rde

import "testing"

func TestVNCCredentialsFromSession_DecomposesHostPort(t *testing.T) {
	for _, tc := range []struct {
		name     string
		addr     string
		wantHost string
		wantPort int
	}{
		{name: "host:port", addr: "host.example:5901", wantHost: "host.example", wantPort: 5901},
		{name: "vnc:// prefix", addr: "vnc://host.example:5900", wantHost: "host.example", wantPort: 5900},
		{name: "bare host defaults to 5900", addr: "host.example", wantHost: "host.example", wantPort: 5900},
		// The host/port split is shared with parseSSHAddress, so an IPv6
		// endpoint survives instead of being truncated at its first colon.
		{name: "bracketed IPv6 with port", addr: "[2001:db8::1]:5901", wantHost: "2001:db8::1", wantPort: 5901},
		{name: "bare IPv6 defaults to 5900", addr: "vnc://2001:db8::1", wantHost: "2001:db8::1", wantPort: 5900},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := VNCCredentialsFromSession(Session{Status: "running", VNCAddress: tc.addr})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tc.wantHost || got.Port != tc.wantPort {
				t.Errorf("host/port = %q/%d, want %q/%d", got.Host, got.Port, tc.wantHost, tc.wantPort)
			}
		})
	}
}

// TestFormatVNCURL pins the userinfo encoding. url.QueryEscape is form
// encoding: it emits "+" for a space and "%2B" for a literal "+", so a client
// decodes either back to the wrong password and auth silently fails.
func TestFormatVNCURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		host string
		port int
		user string
		pass string
		want string
	}{
		{name: "no credentials", host: "h", port: 5900, want: "vnc://h:5900"},
		{name: "user only", host: "h", port: 5900, user: "vagrant", want: "vnc://vagrant@h:5900"},
		{name: "password only", host: "h", port: 5900, pass: "hunter2", want: "vnc://:hunter2@h:5900"},
		{name: "plain credentials", host: "h", port: 5901, user: "u", pass: "p", want: "vnc://u:p@h:5901"},
		{name: "at and colon are escaped", host: "h", port: 5900, user: "user@x", pass: "a:b@c", want: "vnc://user%40x:a%3Ab%40c@h:5900"},
		{name: "space is percent-encoded, not plus", host: "h", port: 5900, user: "u", pass: "my pass", want: "vnc://u:my%20pass@h:5900"},
		{name: "literal plus stays literal", host: "h", port: 5900, user: "u", pass: "p+w", want: "vnc://u:p+w@h:5900"},
		{name: "space and plus together", host: "h", port: 5900, user: "user", pass: "p@ss w+rd", want: "vnc://user:p%40ss%20w+rd@h:5900"},
		{name: "IPv6 host is bracketed", host: "2001:db8::1", port: 5900, user: "u", pass: "p", want: "vnc://u:p@[2001:db8::1]:5900"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatVNCURL(tc.host, tc.port, tc.user, tc.pass); got != tc.want {
				t.Errorf("FormatVNCURL(%q, %d, %q, %q) = %q, want %q", tc.host, tc.port, tc.user, tc.pass, got, tc.want)
			}
		})
	}
}

func TestVNCCredentialsFromSession(t *testing.T) {
	for _, tc := range []struct {
		name      string
		sess      Session
		wantURL   string
		wantError bool
	}{
		{
			name:    "host:port + credentials",
			sess:    Session{Status: "running", VNCAddress: "host.example:5901", VNCUsername: "vagrant", VNCPassword: "hunter2"},
			wantURL: "vnc://vagrant:hunter2@host.example:5901",
		},
		{
			name:    "vnc:// prefix is stripped",
			sess:    Session{Status: "running", VNCAddress: "vnc://host.example:5900", VNCUsername: "u", VNCPassword: "p"},
			wantURL: "vnc://u:p@host.example:5900",
		},
		{
			name:    "bare host defaults to 5900",
			sess:    Session{Status: "running", VNCAddress: "host.example", VNCUsername: "u", VNCPassword: "p"},
			wantURL: "vnc://u:p@host.example:5900",
		},
		{
			name:    "special chars are URL-escaped",
			sess:    Session{Status: "running", VNCAddress: "h:5900", VNCUsername: "user@x", VNCPassword: "a:b@c"},
			wantURL: "vnc://user%40x:a%3Ab%40c@h:5900",
		},
		{
			name:      "no VNC address while running -> not-exposed error",
			sess:      Session{Status: "running"},
			wantError: true,
		},
		{
			name:      "no VNC address + terminated -> status error",
			sess:      Session{Status: "terminated"},
			wantError: true,
		},
		{
			name:      "invalid port",
			sess:      Session{Status: "running", VNCAddress: "h:notaport"},
			wantError: true,
		},
		{
			name:      "zero port",
			sess:      Session{Status: "running", VNCAddress: "h:0"},
			wantError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := VNCCredentialsFromSession(tc.sess)
			if tc.wantError {
				if err == nil {
					t.Fatalf("want error, got url=%q", got.URL)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.URL != tc.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tc.wantURL)
			}
			if got.Address != tc.sess.VNCAddress {
				t.Errorf("Address = %q, want %q", got.Address, tc.sess.VNCAddress)
			}
		})
	}
}
