package goutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePackageNameFromURL(t *testing.T) {
	t.Log("No error - parse package name")
	{
		testListMap := map[string]string{
			"https://github.com/bitrise-io/bitrise.io.git": "github.com/bitrise-io/bitrise.io",
			"git@github.com:bitrise-io/bitrise.io.git":     "github.com/bitrise-io/bitrise.io",
			//
			"https://github.com/bitrise-io/go-utils.git": "github.com/bitrise-io/go-utils",
			"git@github.com:bitrise-io/go-utils.git":     "github.com/bitrise-io/go-utils",
			//
			"https://github.com/my_usr-name.here/repo-part_1.here.git": "github.com/my_usr-name.here/repo-part_1.here",
			"git@github.com:my_usr-name.here/repo-part_1.here.git":     "github.com/my_usr-name.here/repo-part_1.here",
			// no .git
			"https://github.com/my_usr-name.here/repo-part_1.here": "github.com/my_usr-name.here/repo-part_1.here",
			// .git.git
			"https://github.com/my_usr-name.here/repo-part_1.git.git": "github.com/my_usr-name.here/repo-part_1.git",
		}

		for k, v := range testListMap {
			packageName, err := ParsePackageNameFromURL(k)
			require.NoError(t, err)
			require.Equal(t, v, packageName)
		}
	}

	t.Log("Invalid remote URL - parse error")
	{
		testListMap := map[string]string{
			"git@github.com:double:my_usr-name.here/repo-part_1.here.git": "More than one ':' found in the Host part of the URL (git@github.com:double:my_usr-name.here/repo-part_1.here.git)",
			"my_usr-name/repo-part_1.here":                                "No Host found in URL (my_usr-name/repo-part_1.here)",
			"https://github.com":                                          "No Path found in URL (https://github.com)",
			"https://github.com/":                                         "No Path found in URL (https://github.com/)",
		}

		for k, v := range testListMap {
			packageName, err := ParsePackageNameFromURL(k)
			require.EqualError(t, err, v)
			require.Equal(t, "", packageName)
		}
	}
}
