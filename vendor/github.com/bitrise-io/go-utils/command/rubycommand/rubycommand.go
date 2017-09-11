package rubycommand

import (
	"errors"
	"regexp"

	"bufio"
	"bytes"

	"fmt"

	"github.com/bitrise-io/go-utils/command"
)

const (
	systemRubyPth = "/usr/bin/ruby"
	brewRubyPth   = "/usr/local/bin/ruby"
)

// InstallType ...
type InstallType int8

const (
	// Unkown ...
	Unkown InstallType = iota
	// SystemRuby ...
	SystemRuby
	// BrewRuby ...
	BrewRuby
	// RVMRuby ...
	RVMRuby
	// RbenvRuby ...
	RbenvRuby
)

func cmdExist(slice ...string) bool {
	if len(slice) == 0 {
		return false
	}

	cmd, err := command.NewWithParams(slice...)
	if err != nil {
		return false
	}

	return (cmd.Run() == nil)
}

func installType() InstallType {
	whichRuby, err := command.New("which", "ruby").RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return Unkown
	}

	installType := Unkown
	if whichRuby == systemRubyPth {
		installType = SystemRuby
	} else if whichRuby == brewRubyPth {
		installType = BrewRuby
	} else if cmdExist("rvm", "-v") {
		installType = RVMRuby
	} else if cmdExist("rbenv", "-v") {
		installType = RbenvRuby
	}

	return installType
}

func sudoNeeded(installType InstallType, slice ...string) bool {
	if installType != SystemRuby {
		return false
	}

	if len(slice) < 2 {
		return false
	}

	name := slice[0]
	command := slice[1]
	if name == "bundle" {
		return (command == "install" || command == "update")
	} else if name == "gem" {
		return (command == "install" || command == "uninstall")
	}

	return false
}

// NewWithParams ...
func NewWithParams(params ...string) (*command.Model, error) {
	rubyInstallType := installType()
	if rubyInstallType == Unkown {
		return nil, errors.New("unkown ruby installation type")
	}

	if sudoNeeded(rubyInstallType, params...) {
		params = append([]string{"sudo"}, params...)
	}

	return command.NewWithParams(params...)
}

// NewFromSlice ...
func NewFromSlice(slice []string) (*command.Model, error) {
	return NewWithParams(slice...)
}

// New ...
func New(name string, args ...string) (*command.Model, error) {
	slice := append([]string{name}, args...)
	return NewWithParams(slice...)
}

// GemUpdate ...
func GemUpdate(gem string) ([]*command.Model, error) {
	cmds := []*command.Model{}

	cmd, err := New("gem", "update", gem, "--no-document")
	if err != nil {
		return []*command.Model{}, err
	}

	cmds = append(cmds, cmd)

	rubyInstallType := installType()
	if rubyInstallType == RbenvRuby {
		cmd, err := New("rbenv", "rehash")
		if err != nil {
			return []*command.Model{}, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// GemInstall ...
func GemInstall(gem, version string) ([]*command.Model, error) {
	cmds := []*command.Model{}

	slice := []string{"gem", "install", gem, "--no-document"}
	if version != "" {
		slice = append(slice, "-v", version)
	}

	cmd, err := NewFromSlice(slice)
	if err != nil {
		return []*command.Model{}, err
	}

	cmds = append(cmds, cmd)

	rubyInstallType := installType()
	if rubyInstallType == RbenvRuby {
		cmd, err := New("rbenv", "rehash")
		if err != nil {
			return []*command.Model{}, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func findGemInList(gemList, gem, version string) (bool, error) {
	// minitest (5.10.1, 5.9.1, 5.9.0, 5.8.3, 4.7.5)
	pattern := fmt.Sprintf(`^%s \(.*%s.*\)`, gem, version)
	re := regexp.MustCompile(pattern)

	reader := bytes.NewReader([]byte(gemList))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindString(line)
		if match != "" {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// IsGemInstalled ...
func IsGemInstalled(gem, version string) (bool, error) {
	cmd, err := New("gem", "list")
	if err != nil {
		return false, err
	}

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return false, err
	}

	return findGemInList(out, gem, version)
}
