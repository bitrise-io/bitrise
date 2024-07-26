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
			name:   "SSH with user and port",
			gitURL: "ssh://bitrise-bot@github:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "bitrise-bot",
				Host: "github",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "SSH without user but with port",
			gitURL: "ssh://github:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "SSH without port",
			gitURL: "ssh://github/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "SCP-like syntax with user",
			gitURL: "bitrise-bot@github.com:bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "bitrise-bot",
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "SCP-like syntax without user",
			gitURL: "github.com:bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github.com",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "Git protocol with port",
			gitURL: "git://github:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "Git protocol without port",
			gitURL: "git://github/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "HTTPS with port",
			gitURL: "https://github:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "HTTPS without port",
			gitURL: "https://github/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "HTTP with port",
			gitURL: "http://github:22/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "22",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:   "HTTP without port",
			gitURL: "http://github/bitrise-io/bitrise.git",
			want: &GitRepoURL{
				User: "",
				Host: "github",
				Port: "",
				Path: "bitrise-io/bitrise.git",
			},
		},
		{
			name:    "Invalid URL",
			gitURL:  "justastring",
			want:    nil,
			wantErr: "unsupported git URL format",
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
