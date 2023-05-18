package suggestionbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/command/git"
	"gopkg.in/yaml.v2"
)

const (
	suggestionMapRepoURL = "https://github.com/godrei/suggestion-map.git"
	suggestionMapBranch  = "main"
)

var suggestionBox *SuggestionBox

func Setup() error {
	bitriseHome := configs.GetBitriseHomeDirPath()
	suggestionBoxHome := filepath.Join(bitriseHome, "suggestion_box")

	gitRepo, err := git.New(suggestionBoxHome)
	if err != nil {
		return err
	}

	_, err = os.Stat(suggestionBoxHome)
	if err == nil {
		if err := gitRepo.Pull().Run(); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		if err := gitRepo.CloneTagOrBranch(suggestionMapRepoURL, suggestionMapBranch).Run(); err != nil {
			return err
		}
	} else {
		return err
	}

	suggestionMapPth := filepath.Join(suggestionBoxHome, "suggestions.yaml")

	suggestionBox, err = loadFile(suggestionMapPth)
	if err != nil {
		return err
	}

	return nil
}

func AddSuggestion(err error, context string) error {
	errorMessage := err.Error()
	suggestions := findSuggestions(context, errorMessage)
	if len(suggestions) == 0 {
		return err
	}

	return fmt.Errorf("%s\nSuggestions:\n%s", errorMessage, strings.Join(suggestions, "\n"))
}

func findSuggestions(context string, input string) []string {
	if suggestionBox == nil {
		return nil
	}

	tests := suggestionBox.matchers[context]
	if len(tests) == 0 {
		return nil
	}

	suggestions := make([]string, 0)
	for _, test := range tests {
		re := regexp.MustCompile(test.MatchRegex)
		if !re.MatchString(input) {
			continue
		}

		suggestion := re.ReplaceAllString(input, test.ReplacementPattern)
		if suggestion == "" {
			continue
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

type RegularError struct {
	MatchRegex         string `json:"match,omitempty" yaml:"match"`
	ReplacementPattern string `json:"replacement,omitempty" yaml:"replacement"`
}

type SuggestionBox struct {
	matchers map[string][]RegularError
}

func loadFile(path string) (*SuggestionBox, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return load(bytes)
}

func load(bytes []byte) (*SuggestionBox, error) {
	var matchers map[string][]RegularError
	err := yaml.Unmarshal(bytes, &matchers)
	if err != nil {
		return nil, err
	}
	return &SuggestionBox{
		matchers: matchers,
	}, nil
}
