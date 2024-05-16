package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/envman/models"
)

// Action is a possible action changing an environment variable
type Action int

const (
	// InvalidAction represents an unexpected state
	InvalidAction Action = iota + 1
	// SetAction is an environment variable assignment, like os.Setenv
	SetAction
	// UnsetAction is an action to clear (if existing) an environment variable, like os.Unsetenv
	UnsetAction
	// SkipAction means that no action is performed (usually for an env with an empty value)
	SkipAction
)

// Command describes an action performed on an environment variable
type Command struct {
	Action   Action
	Variable Variable
}

// Variable is an environment variable
type Variable struct {
	Key   string
	Value string
}

// DeclarationSideEffects is returned by GetDeclarationsSideEffects()
type DeclarationSideEffects struct {
	// CommandHistory is an ordered list of commands: when performed in sequence,
	// will result in a environment that contains the declared env vars
	CommandHistory []Command
	// ResultEnvironment is returned for reference,
	// it will equal the environment after performing the commands
	ResultEnvironment map[string]string
	// EvaluatedNewEnvs is the set of envs resulted after evaluating newEnvs with envSource
	EvaluatedNewEnvs map[string]string
}

// EnvironmentSource implementations can return an initial environment
type EnvironmentSource interface {
	GetEnvironment() map[string]string
}

// DefaultEnvironmentSource is a default implementation of EnvironmentSource, returns the current environment
type DefaultEnvironmentSource struct{}

// GetEnvironment returns the current process' environment
func (*DefaultEnvironmentSource) GetEnvironment() map[string]string {
	processEnvs := os.Environ()
	envs := make(map[string]string)

	// String names can be duplicated (on Unix), and the Go libraries return the first instance of them:
	// https://github.com/golang/go/blob/98d20fb23551a7ab900fcfe9d25fd9cb6a98a07f/src/syscall/env_unix.go#L45
	// From https://pubs.opengroup.org/onlinepubs/9699919799/:
	// > "There is no meaning associated with the order of strings in the environment.
	// > If more than one string in an environment of a process has the same name, the consequences are undefined."
	for _, env := range processEnvs {
		key, value := SplitEnv(env)
		if key == "" {
			continue
		}

		envs[key] = value
	}

	return envs
}

// SplitEnv splits an env returned by os.Environ
func SplitEnv(env string) (key string, value string) {
	const sep = "="
	split := strings.SplitAfterN(env, sep, 2)
	if split == nil {
		return "", ""
	}
	key = strings.TrimSuffix(split[0], sep)
	if len(split) > 1 {
		value = split[1]
	}
	return
}

// GetDeclarationsSideEffects iterates over the list of ordered new declared variables sequentially and returns the needed
// commands (like os.Setenv) to add the variables to the current environment.
// The current process environment is not changed.
// Variable expansion is done also, every new variable can reference the previous and initial environments (via EnvironmentSource)
// The new variables (models.EnvironmentItemModel) can be defined in the envman definition file, or filled in directly.
// If the source of the variables (models.EnvironmentItemModel) is the bitrise.yml workflow,
// they will be in this order:
//  - Bitrise CLI configuration parameters (IS_CI, IS_DEBUG)
//  - App secrets
//  - App level envs
//  - Workflow level envs
//  - Additional Step inputs envs (BITRISE_STEP_SOURCE_DIR; BitriseTestDeployDirEnvKey ("BITRISE_TEST_DEPLOY_DIR"), PWD)
//  - Input envs
func GetDeclarationsSideEffects(newEnvs []models.EnvironmentItemModel, envSource EnvironmentSource) (DeclarationSideEffects, error) {
	envs := envSource.GetEnvironment()
	commandHistory := make([]Command, len(newEnvs))
	evaluatedNewEnvs := make(map[string]string, len(newEnvs))

	for i, env := range newEnvs {
		command, err := getDeclarationCommand(env, envs)
		if err != nil {
			return DeclarationSideEffects{}, fmt.Errorf("failed to parse new environment variable (%s): %s", env, err)
		}

		commandHistory[i] = command

		switch command.Action {
		case SetAction:
			envs[command.Variable.Key] = command.Variable.Value
			evaluatedNewEnvs[command.Variable.Key] = command.Variable.Value
		case UnsetAction:
			delete(envs, command.Variable.Key)
			delete(evaluatedNewEnvs, command.Variable.Key)
		case SkipAction:
		default:
			return DeclarationSideEffects{}, fmt.Errorf("invalid case for environment declaration action: %#v", command)
		}
	}

	return DeclarationSideEffects{
		CommandHistory:    commandHistory,
		ResultEnvironment: envs,
		EvaluatedNewEnvs:  evaluatedNewEnvs,
	}, nil
}

// getDeclarationCommand maps a variable to be declared (env) to an expanded env key and value.
// The current process environment is not changed.
func getDeclarationCommand(env models.EnvironmentItemModel, envs map[string]string) (Command, error) {
	envKey, envValue, err := env.GetKeyValuePair()
	if err != nil {
		return Command{}, fmt.Errorf("failed to get new environment variable name and value: %s", err)
	}

	options, err := env.GetOptions()
	if err != nil {
		return Command{}, fmt.Errorf("failed to get new environment options: %s", err)
	}

	if options.Unset != nil && *options.Unset {
		return Command{
			Action:   UnsetAction,
			Variable: Variable{Key: envKey},
		}, nil
	}

	if options.SkipIfEmpty != nil && *options.SkipIfEmpty && envValue == "" {
		return Command{
			Action:   SkipAction,
			Variable: Variable{Key: envKey},
		}, nil
	}

	mappingFuncFactory := func(envs map[string]string) func(string) string {
		return func(key string) string {
			if _, ok := envs[key]; !ok {
				return ""
			}

			return envs[key]
		}
	}

	if options.IsExpand != nil && *options.IsExpand {
		envValue = os.Expand(envValue, mappingFuncFactory(envs))
	}

	return Command{
		Action: SetAction,
		Variable: Variable{
			Key:   envKey,
			Value: envValue,
		},
	}, nil
}
