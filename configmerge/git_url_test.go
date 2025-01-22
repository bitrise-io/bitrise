package configmerge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseGitURL(t *testing.T) {
	tests := []struct {
		name    string
		gitURL  string
		want    *GitRepoURL
		wantErr string
	}{
		{
			name:   "SCP-like SSH syntax",
			gitURL: "git@github.com:bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User:           "git",
				Host:           "github.com",
				Path:           "bitrise-io/bitrise.git",
				OriginalSyntax: SSHGitRepoURLSyntax,
			},
		},
		{
			name:   "SCP-like SSH syntax without user",
			gitURL: "github.com:bitrise-io/bitrise.git",
			want: &GitRepoURL{
				Host:           "github.com",
				Path:           "bitrise-io/bitrise.git",
				OriginalSyntax: SSHGitRepoURLSyntax,
			},
		},
		{
			name:   "HTTPS syntax",
			gitURL: "https://github.com/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				Host:           "github.com",
				Path:           "bitrise-io/bitrise.git",
				OriginalSyntax: HTTPSRepoURLSyntax,
			},
		},
		{
			name:   "HTTPS syntax with port",
			gitURL: "https://github.com:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User:           "",
				Host:           "github.com",
				Port:           "22",
				Path:           "bitrise-io/bitrise.git",
				OriginalSyntax: HTTPSRepoURLSyntax,
			},
		},
		{
			name:    "Invalid URL",
			gitURL:  "justastring",
			want:    nil,
			wantErr: "unsupported git URL format: justastring",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitRepoURL(tt.gitURL)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGitRepoURL_URLString(t *testing.T) {
	tests := []struct {
		name       string
		gitRepoURL GitRepoURL
		syntax     GitRepoURLSyntax
		want       string
	}{
		{
			name: "git repo url to SSH syntax",
			gitRepoURL: GitRepoURL{
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: SSHGitRepoURLSyntax,
			want:   "github.com:bitrise-io/bitrise.git",
		},
		{
			name: "git repo url (with user) to SSH syntax",
			gitRepoURL: GitRepoURL{
				User: "git",
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: SSHGitRepoURLSyntax,
			want:   "git@github.com:bitrise-io/bitrise.git",
		},
		{
			name: "git repo url (with user and port) to SSH syntax",
			gitRepoURL: GitRepoURL{
				User: "git",
				Host: "github.com",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: SSHGitRepoURLSyntax,
			want:   "git@github.com:bitrise-io/bitrise.git",
		},
		{
			name: "git repo url to HTTPS syntax",
			gitRepoURL: GitRepoURL{
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: HTTPSRepoURLSyntax,
			want:   "https://github.com/bitrise-io/bitrise.git",
		},
		{
			name: "git repo url (with user) to HTTPS syntax",
			gitRepoURL: GitRepoURL{
				User: "git",
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: HTTPSRepoURLSyntax,
			want:   "https://github.com/bitrise-io/bitrise.git",
		},
		{
			name: "git repo url (with user and port) to HTTPS syntax",
			gitRepoURL: GitRepoURL{
				User: "git",
				Host: "github.com",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
			syntax: HTTPSRepoURLSyntax,
			want:   "https://github.com:22/bitrise-io/bitrise.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gitRepoURL.URLString(tt.syntax)
			require.Equal(t, tt.want, got)
		})
	}
}
