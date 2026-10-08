package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestReleasedVersionsInLine(t *testing.T) {
	t.Run("reverses mise's order without sorting", func(t *testing.T) {
		remote := []string{"1.19.5-otp-28", "1.20.4", "1.20.4-otp-28", "1.20.4-otp-29"}

		versions := releasedVersionsInLine(remote, "elixir", "")
		assert.Equal(t, []string{"1.20.4-otp-29", "1.20.4-otp-28", "1.20.4", "1.19.5-otp-28"}, versions)
	})

	t.Run("keeps the prefix's line where mise continues it", func(t *testing.T) {
		tests := []struct {
			tool     provider.ToolID
			prefix   string
			versions []string
			want     []string
		}{
			{"nodejs", "20", []string{"18.0.0", "20.0.0", "20.1.0", "22.0.0"}, []string{"20.1.0", "20.0.0"}},
			{"nodejs", "18.1", []string{"18.1.0", "18.1.2", "18.10.0"}, []string{"18.1.2", "18.1.0"}},
			{"nodejs", "22", []string{"18.0.0", "20.0.0"}, nil},
			{"nodejs", "20.", []string{"20.0.0", "20.1.0", "200.0.0"}, []string{"20.1.0", "20.0.0"}},
			{"nodejs", ".", []string{"18.0.0", "20.0.0"}, []string{"20.0.0", "18.0.0"}},
			{"nodejs", "22", []string{"22-", "22.1.0"}, []string{"22.1.0"}},
			{"elixir", "1.19.0", []string{"1.19.0-otp-27", "1.19.0-otp-28", "1.19.00"}, []string{"1.19.0-otp-28", "1.19.0-otp-27"}},
			{"flutter", "1.7.8", []string{"1.7.8+hotfix.4-stable", "1.7.80"}, []string{"1.7.8+hotfix.4-stable"}},
			{"ruby", "truffleruby", []string{"truffleruby-24.1.0", "truffleruby+graalvm-24.1.0"}, []string{"truffleruby-24.1.0"}},
			{"nodejs", "v22", []string{"22.1.0", "v22.0.0", "220.0.0"}, []string{"v22.0.0", "22.1.0"}},
			{"golang", "1.22", []string{"v1.22.0", "1.220.0"}, []string{"v1.22.0"}},
			{"java", "temurin-21.0.5", []string{"temurin-21.0.5+11.0.LTS", "temurin-21.0.50"}, []string{"temurin-21.0.5+11.0.LTS"}},
			{"java", "temurin-", []string{"temurin-21.0.5+11.0.LTS", "zulu-21.52.203.0"}, []string{"temurin-21.0.5+11.0.LTS"}},
			{"java", "21", []string{"21.0.2", "v21.0.1"}, []string{"21.0.2"}},
		}
		for _, tt := range tests {
			versions := releasedVersionsInLine(tt.versions, tt.tool, tt.prefix)
			assert.Equal(t, tt.want, versions, "%s@%s", tt.tool, tt.prefix)
		}
	})
}
