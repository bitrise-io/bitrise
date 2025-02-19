package toolversions

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/bitrise/v2/log"
)

type ToolVersionReporter interface {
	// IsAvailable returns true if the tool version manager is available and actively manages the tool versions.
	IsAvailable() bool

	// CurrentToolVersions returns a snapshot of the currently active tools and versions.
	// The returned map is keyed by tool name.
	// Tool names and reported versions are implementation-specific. Tool names are normalized to lowercase.
	CurrentToolVersions() (map[string]ToolVersion, error)
}

type ToolVersion struct {
	Version        string `json:"version"`
	IsInstalled    bool   `json:"is_installed"`
	DeclaredByFile string `json:"declared_by_file"`
	IsGlobal       bool   `json:"is_global"`
}

type ASDFVersionReporter struct {
	cmdLocator  env.CommandLocator
	cmdFactory  command.Factory
	logger      log.Logger
	userHomeDir string
}

func NewASDFVersionReporter(cmdLocator env.CommandLocator, cmdFactory command.Factory, logger log.Logger, userHomeDir string) ASDFVersionReporter {
	return ASDFVersionReporter{
		cmdLocator:  cmdLocator,
		cmdFactory:  cmdFactory,
		logger:      logger,
		userHomeDir: userHomeDir,
	}
}

func (r *ASDFVersionReporter) IsAvailable() bool {
	_, err := r.cmdLocator.LookPath("asdf")
	if err != nil {
		r.logger.Debugf("asdf not found in path")
		return false
	}

	code, err := r.cmdFactory.Create("asdf", []string{"current"}, &command.Opts{}).RunAndReturnExitCode()
	if err != nil {
		r.logger.Debugf("run asdf current: %s", err)
		return false
	}
	if code != 0 {
		r.logger.Debugf("run asdf current: nonzero exit code: %d", code)
		return false
	}

	return true
}

func (r *ASDFVersionReporter) CurrentToolVersions() (map[string]ToolVersion, error) {
	cmd := r.cmdFactory.Create("asdf", []string{"current"}, &command.Opts{})
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run asdf current: %s", err)
	}

	asdfVersionRegexp, err := regexp.Compile(`([a-z]+)\s+(\S+)\s+(.+)`)
	if err != nil {
		return nil, fmt.Errorf("compile regex: %s", err)
	}

	toolVersions := map[string]ToolVersion{}
	for _, line := range strings.Split(out, "\n") {
		matches := asdfVersionRegexp.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		captureGroups := matches[0]
		if len(captureGroups) != 4 {
			// Entire match + 3 capture groups
			return nil, fmt.Errorf("unexpected number of matches (%d) in input: %s, matches: %s", len(matches), line, matches)
		}
		tool := captureGroups[1]
		version := captureGroups[2]
		declaredBy := captureGroups[3]

		if tool == "alias" {
			// Meta-tool, ignore
			continue
		}

		if version == "______" {
			// No version is set globally, ignore
			continue
		}

		isInstalled := !strings.HasPrefix(declaredBy, "Not installed.")
		var declaredByFile string
		var isGlobal bool
		file := filepath.Base(declaredBy)
		if file != "." && isInstalled {
			declaredByFile = file
			isGlobal = filepath.Dir(declaredBy) == r.userHomeDir
		}

		toolVersions[strings.ToLower(tool)] = ToolVersion{
			Version:        version,
			IsInstalled:    isInstalled,
			DeclaredByFile: declaredByFile,
			IsGlobal:       isGlobal,
		}
	}

	return toolVersions, nil
}
