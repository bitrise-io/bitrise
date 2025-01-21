package cli

import (
	"encoding/base64"
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func TestMinimalValidation(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		validation bitrise.ValidationType
		wantErr    string
	}{
		{
			name: "Valid config",
			config: `
format_version: '17'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a] }
      c: { uses: a }
workflows:
  a: {}
  b: {}
`,
			validation: bitrise.ValidationTypeFull,
		},
		{
			name: "Only valid with minimal validation",
			config: `
format_version: '17'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a] }
      c: {}
workflows:
  a: {}
  b: {}`,
			validation: bitrise.ValidationTypeMinimal,
		},
		{
			name: "Invalid config even with minimal validation",
			config: `
format_version: '17'
pipelines:
  dag:
    workflows:
      a: {}
      b: { depends_on: [a] }
workflows:
  a: 
    before_run: [c]
  b: {}
`,
			validation: bitrise.ValidationTypeMinimal,
			wantErr:    "failed to get Bitrise config (bitrise.yml) from base 64 data: Failed to parse bitrise config, error: Workflow does not exist with name c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b64Data := base64.StdEncoding.EncodeToString([]byte(tt.config))
			_, warnings, err := CreateBitriseConfigFromCLIParams(b64Data, "", tt.validation)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, 0, len(warnings))
			}
		})
	}
}
