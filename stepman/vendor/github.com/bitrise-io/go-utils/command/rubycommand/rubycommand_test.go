package rubycommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdExist(t *testing.T) {
	t.Log("exist")
	{
		require.Equal(t, true, cmdExist("ls"))
	}

	t.Log("not exist")
	{
		require.Equal(t, false, cmdExist("__not_existing_command__"))
	}
}

func TestSudoNeeded(t *testing.T) {
	t.Log("sudo NOT need")
	{
		require.Equal(t, false, sudoNeeded(Unkown, "ls"))
		require.Equal(t, false, sudoNeeded(SystemRuby, "ls"))
		require.Equal(t, false, sudoNeeded(BrewRuby, "ls"))
		require.Equal(t, false, sudoNeeded(RVMRuby, "ls"))
		require.Equal(t, false, sudoNeeded(RbenvRuby, "ls"))
	}

	t.Log("sudo needed for SystemRuby in case of gem list management command")
	{
		require.Equal(t, false, sudoNeeded(Unkown, "gem", "install", "fastlane"))
		require.Equal(t, true, sudoNeeded(SystemRuby, "gem", "install", "fastlane"))
		require.Equal(t, false, sudoNeeded(BrewRuby, "gem", "install", "fastlane"))
		require.Equal(t, false, sudoNeeded(RVMRuby, "gem", "install", "fastlane"))
		require.Equal(t, false, sudoNeeded(RbenvRuby, "gem", "install", "fastlane"))

		require.Equal(t, false, sudoNeeded(Unkown, "gem", "uninstall", "fastlane"))
		require.Equal(t, true, sudoNeeded(SystemRuby, "gem", "uninstall", "fastlane"))
		require.Equal(t, false, sudoNeeded(BrewRuby, "gem", "uninstall", "fastlane"))
		require.Equal(t, false, sudoNeeded(RVMRuby, "gem", "uninstall", "fastlane"))
		require.Equal(t, false, sudoNeeded(RbenvRuby, "gem", "uninstall", "fastlane"))

		require.Equal(t, false, sudoNeeded(Unkown, "bundle", "install"))
		require.Equal(t, true, sudoNeeded(SystemRuby, "bundle", "install"))
		require.Equal(t, false, sudoNeeded(BrewRuby, "bundle", "install"))
		require.Equal(t, false, sudoNeeded(RVMRuby, "bundle", "install"))
		require.Equal(t, false, sudoNeeded(RbenvRuby, "bundle", "install"))

		require.Equal(t, false, sudoNeeded(Unkown, "bundle", "update"))
		require.Equal(t, true, sudoNeeded(SystemRuby, "bundle", "update"))
		require.Equal(t, false, sudoNeeded(BrewRuby, "bundle", "update"))
		require.Equal(t, false, sudoNeeded(RVMRuby, "bundle", "update"))
		require.Equal(t, false, sudoNeeded(RbenvRuby, "bundle", "update"))
	}
}

func TestFindGemInList(t *testing.T) {
	t.Log("finds gem")
	{
		gemList := `
*** LOCAL GEMS ***

addressable (2.5.0, 2.4.0, 2.3.8)
activesupport (5.0.0.1, 4.2.7.1, 4.2.6, 4.2.5, 4.1.16, 4.0.13)
angularjs-rails (1.5.8)`

		found, err := findGemInList(gemList, "activesupport", "")
		require.NoError(t, err)
		require.Equal(t, true, found)
	}

	t.Log("finds gem with version")
	{
		gemList := `
*** LOCAL GEMS ***

addressable (2.5.0, 2.4.0, 2.3.8)
activesupport (5.0.0.1, 4.2.7.1, 4.2.6, 4.2.5, 4.1.16, 4.0.13)
angularjs-rails (1.5.8)`

		found, err := findGemInList(gemList, "activesupport", "4.2.5")
		require.NoError(t, err)
		require.Equal(t, true, found)
	}

	t.Log("gem version not found in list")
	{
		gemList := `
*** LOCAL GEMS ***

addressable (2.5.0, 2.4.0, 2.3.8)
activesupport (5.0.0.1, 4.2.7.1, 4.2.6, 4.2.5, 4.1.16, 4.0.13)
angularjs-rails (1.5.8)`

		found, err := findGemInList(gemList, "activesupport", "0.9.0")
		require.NoError(t, err)
		require.Equal(t, false, found)
	}

	t.Log("gem not found in list")
	{
		gemList := `
*** LOCAL GEMS ***

addressable (2.5.0, 2.4.0, 2.3.8)
activesupport (5.0.0.1, 4.2.7.1, 4.2.6, 4.2.5, 4.1.16, 4.0.13)
angularjs-rails (1.5.8)`

		found, err := findGemInList(gemList, "fastlane", "")
		require.NoError(t, err)
		require.Equal(t, false, found)
	}

	t.Log("gem with version not found in list")
	{
		gemList := `
*** LOCAL GEMS ***

addressable (2.5.0, 2.4.0, 2.3.8)
activesupport (5.0.0.1, 4.2.7.1, 4.2.6, 4.2.5, 4.1.16, 4.0.13)
angularjs-rails (1.5.8)`

		found, err := findGemInList(gemList, "fastlane", "2.70")
		require.NoError(t, err)
		require.Equal(t, false, found)
	}
}
