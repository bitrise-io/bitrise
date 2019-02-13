package rubyscript

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

const gemfileContent = `# frozen_string_literal: true
source "https://rubygems.org"

gem "json"
`

const gemfileLockContent = `GEM
  remote: https://rubygems.org/
  specs:
    json (2.1.0)

PLATFORMS
  ruby

DEPENDENCIES
  json

BUNDLED WITH
   1.15.3
`

const rubyScriptWithGemContent = `require 'json'

begin
  messageObj = '{"message":"Hi Bitrise"}'
  messageJSON = JSON.parse(messageObj)
  puts "#{{ :data =>  messageJSON['message'] }.to_json}"
rescue => e
	puts "#{{ :error => e.to_s }.to_json}"
end
`

const rubyScriptContent = `puts '{"data":"Hi Bitrise"}'`

func TestNew(t *testing.T) {
	t.Log("initialize new ruby script runner with the ruby script content")
	{
		runner := New(rubyScriptContent)
		require.NotNil(t, runner)
	}
}

func Test_ensureTmpDir(t *testing.T) {
	t.Log("ensure runner holds a tmp dir path")
	{
		runner := New(rubyScriptContent)
		require.NotNil(t, runner)

		tmpDir, err := runner.ensureTmpDir()
		require.NoError(t, err)

		exist, err := pathutil.IsDirExists(tmpDir)
		require.NoError(t, err)
		require.True(t, exist)
	}
}

func TestBundleInstallCommand(t *testing.T) {
	t.Log("bundle install gems")
	{
		runner := New(rubyScriptWithGemContent)
		require.NotNil(t, runner)

		bundleInstallCmd, err := runner.BundleInstallCommand(gemfileContent, gemfileLockContent)
		require.NoError(t, err)

		cmd := bundleInstallCmd.GetCmd()

		require.Equal(t, filepath.Base(cmd.Path), "bundle")
		require.Equal(t, 3, len(cmd.Args))
		require.Equal(t, "bundle", cmd.Args[0])
		require.Equal(t, "install", cmd.Args[1])

		gemfileFlag := cmd.Args[2]
		split := strings.Split(gemfileFlag, "=")
		require.Equal(t, 2, len(split))
		require.Equal(t, "--gemfile", split[0])
		require.Equal(t, "Gemfile", filepath.Base(split[1]))

		require.NoError(t, bundleInstallCmd.Run())
	}
}

func TestRunScriptCommand(t *testing.T) {
	t.Log("runs 'ruby script.rb'")
	{
		runner := New(rubyScriptContent)
		require.NotNil(t, runner)

		runCmd, err := runner.RunScriptCommand()
		require.NoError(t, err)

		cmd := runCmd.GetCmd()

		require.Equal(t, filepath.Base(cmd.Path), "ruby")
		require.Equal(t, 2, len(cmd.Args))
		require.Equal(t, "ruby", cmd.Args[0])
		require.Equal(t, "script.rb", filepath.Base(cmd.Args[1]))

		out, err := runCmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.Equal(t, `{"data":"Hi Bitrise"}`, out)
	}

	t.Log("runy 'bundle exec ruby script.rb', if Gemfile installed with bundler")
	{
		runner := New(rubyScriptWithGemContent)
		require.NotNil(t, runner)

		bundleInstallCmd, err := runner.BundleInstallCommand(gemfileContent, gemfileLockContent)
		require.NoError(t, err)
		t.Logf("$ %s", bundleInstallCmd.PrintableCommandArgs())
		require.NoError(t, bundleInstallCmd.Run())

		runCmd, err := runner.RunScriptCommand()
		require.NoError(t, err)

		cmd := runCmd.GetCmd()

		require.Equal(t, filepath.Base(cmd.Path), "bundle")
		require.Equal(t, 4, len(cmd.Args))
		require.Equal(t, "bundle", cmd.Args[0])
		require.Equal(t, "exec", cmd.Args[1])
		require.Equal(t, "ruby", cmd.Args[2])
		require.Equal(t, "script.rb", filepath.Base(cmd.Args[3]))

		t.Logf("$ %s", runCmd.PrintableCommandArgs())
		out, err := runCmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"data":"Hi Bitrise"}`, out)
	}
}
