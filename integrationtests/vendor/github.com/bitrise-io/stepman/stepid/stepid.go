package stepid

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bitrise-io/stepman/models"
)

// CanonicalID is a structured representation of a composite-step-id
// A composite step id is: step-lib-source::step-id@1.0.0
type CanonicalID struct {
	// steplib source uri, or in case of local path just "path", and in case of direct git url just "git"
	SteplibSource string
	// IDOrURI : ID if steplib is provided, URI if local step or in case a direct git url provided
	IDorURI string
	// Version : version in the steplib, or in case of a direct git step the tag-or-branch to use
	Version string
}

// compositeVersionStr examples:
//   - local path:
//   - path::~/path/to/step/dir
//   - direct git url and branch or tag:
//   - git::https://github.com/bitrise-io/steps-timestamp.git@master
//   - Steplib independent step:
//   - _::https://github.com/bitrise-io/steps-bash-script.git@2.0.0:
//   - full ID with steplib, stepid and version:
//   - https://github.com/bitrise-io/bitrise-steplib.git::script@2.0.0
//   - only stepid and version (requires a default steplib source to be provided):
//   - script@2.0.0
//   - only stepid, latest version will be used (requires a default steplib source to be provided):
//   - script
func CreateCanonicalIDFromString(compositeVersionStr, defaultStepLibSource string) (CanonicalID, error) {
	src := getStepSource(compositeVersionStr)
	if src == "" {
		if defaultStepLibSource == "" {
			return CanonicalID{}, errors.New("no default StepLib source, in this case the composite ID should contain the source, separated with a '::' separator from the step ID (" + compositeVersionStr + ")")
		}
		src = defaultStepLibSource
	}

	id := getStepID(compositeVersionStr)
	if id == "" {
		return CanonicalID{}, errors.New("no ID found at all (" + compositeVersionStr + ")")
	}

	version := getStepVersion(compositeVersionStr)

	return CanonicalID{
		IDorURI:       id,
		SteplibSource: string(src),
		Version:       version,
	}, nil
}

func Validate(compositeVersionString string) error {
	ver := getStepVersion(compositeVersionString)
	src := getStepSource(compositeVersionString)

	if ver != "" && isStepLibSource(src) {
		if _, err := models.ParseRequiredVersion(ver); err != nil {
			return fmt.Errorf("invalid version format (%s) specified for step ID: %s", ver, compositeVersionString)
		}
	}

	return nil
}

// IsUniqueResourceID : true if this ID is a unique resource ID, which is true
// if the ID refers to the exact same step code/data every time.
// Practically, this is only true for steps from StepLibrary collections,
// a local path or direct git step ID is never guaranteed to identify the
// same resource every time, the step's behaviour can change at every execution!
//
// __If the ID is a Unique Resource ID then the step can be cached (locally)__,
// as it won't change between subsequent step execution.
func (sIDData CanonicalID) IsUniqueResourceID() bool {
	if !isStepLibSource(sIDData.SteplibSource) {
		return false
	}

	// in any other case, it's a StepLib URL
	// but it's only unique if StepID and Step Version are all defined!
	if len(sIDData.IDorURI) > 0 && len(sIDData.Version) > 0 {
		return true
	}

	// in every other case, it's not unique, not even if it's from a StepLib
	return false
}

// returns true if step source is StepLib
func isStepLibSource(source string) bool {
	switch source {
	case "path", "git", "_", "":
		return false
	default:
		return true
	}
}

// returns step version from compositeString
// e.g.: "git::https://github.com/bitrise-steplib/steps-script.git@master" -> "master"
func getStepVersion(compositeVersionStr string) string {
	composite := getStepComposite(compositeVersionStr)

	if s := splitCompositeComponents(composite); len(s) > 1 {
		return s[len(s)-1]
	}

	return ""
}

// detaches source from the step node
// e.g.: "git::git@github.com:bitrise-steplib/steps-script.git@master" -> "git"
func getStepSource(compositeVersionStr string) string {
	if s := strings.SplitN(string(compositeVersionStr), "::", 2); len(s) == 2 {
		if src := s[0]; len(src) > 0 {
			return src
		}
	}
	return ""
}

// returns step ID from compositeString
// e.g.: "git::https://github.com/bitrise-steplib/steps-script.git@master" -> "https://github.com/bitrise-steplib/steps-script.git"
func getStepID(compositeVersionStr string) string {
	composite := getStepComposite(compositeVersionStr)
	return splitCompositeComponents(composite)[0]
}

// splits step node composite into it's parts by taking care of extra "@" when using SSH git URL
// e.g.: "git::git@github.com:bitrise-steplib/steps-script.git@master" -> ["git@github.com:bitrise-steplib/steps-script.git" "master"]
func splitCompositeComponents(composite string) []string {
	s := strings.Split(composite, "@")
	if item := s[0]; item == "git" {
		s = s[1:]
		s[0] = item + "@" + s[0]
	}
	return s
}

// detaches step id and version composite from the step node
// e.g.: "git::git@github.com:bitrise-steplib/steps-script.git@master" -> "git@github.com:bitrise-steplib/steps-script.git@master"
func getStepComposite(compositeVersionStr string) string {
	if s := strings.SplitN(compositeVersionStr, "::", 2); len(s) == 2 {
		return s[1]
	}
	return compositeVersionStr
}
