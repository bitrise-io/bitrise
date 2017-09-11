package rubyscript

import (
	"path"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// Helper ...
type Helper struct {
	scriptContent string
	tmpDir        string
	gemfilePth    string
}

// New ...
func New(scriptContent string) Helper {
	return Helper{
		scriptContent: scriptContent,
	}
}

func (h *Helper) ensureTmpDir() (string, error) {
	if h.tmpDir != "" {
		return h.tmpDir, nil
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__ruby-script-runner__")
	if err != nil {
		return "", err
	}

	h.tmpDir = tmpDir

	return tmpDir, nil
}

// BundleInstallCommand ...
func (h *Helper) BundleInstallCommand(gemfileContent, gemfileLockContent string) (*command.Model, error) {
	tmpDir, err := h.ensureTmpDir()
	if err != nil {
		return nil, err
	}

	gemfilePth := path.Join(tmpDir, "Gemfile")
	if err := fileutil.WriteStringToFile(gemfilePth, gemfileContent); err != nil {
		return nil, err
	}

	if gemfileLockContent != "" {
		gemfileLockPth := path.Join(tmpDir, "Gemfile.lock")
		if err := fileutil.WriteStringToFile(gemfileLockPth, gemfileLockContent); err != nil {
			return nil, err
		}
	}

	h.gemfilePth = gemfilePth

	// use '--gemfile=<gemfile>' flag to specify Gemfile path
	// ... In general, bundler will assume that the location of the Gemfile(5) is also the project root,
	// and will look for the Gemfile.lock and vendor/cache relative to it. ...
	return command.New("bundle", "install", "--gemfile="+gemfilePth), nil
}

// RunScriptCommand ...
func (h Helper) RunScriptCommand() (*command.Model, error) {
	tmpDir, err := h.ensureTmpDir()
	if err != nil {
		return nil, err
	}

	rubyScriptPth := path.Join(tmpDir, "script.rb")
	if err := fileutil.WriteStringToFile(rubyScriptPth, h.scriptContent); err != nil {
		return nil, err
	}

	var cmd *command.Model
	if h.gemfilePth != "" {
		cmd = command.New("bundle", "exec", "ruby", rubyScriptPth)
		cmd.AppendEnvs("BUNDLE_GEMFILE=" + h.gemfilePth)
	} else {
		cmd = command.New("ruby", rubyScriptPth)
	}

	return cmd, nil
}
