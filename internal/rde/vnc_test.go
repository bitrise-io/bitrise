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
