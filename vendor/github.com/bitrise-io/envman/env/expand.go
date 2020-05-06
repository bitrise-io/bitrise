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
	// SetAction is an environment variable assignement, like os.Setenv
	SetAction
	// UnsetAction is an action to clear (if existing) an environment variable, like os.Unsetenv
	UnsetAction
	// SkipAction means that no action is performed (usually for an env with an empty value)
	SkipAction
)

// Command describes an action performed on an envrionment variable
type Command struct {
	Action   Action
	Variable Variable
}

// Variable is an environment variable
type Variable struct {
	Key   string
	Value string
	// IsSensitive is true if variable is marked (optionally) sensitive initally (for example a sensitive input or a secret),
	// or recursively references any variable marked as sensitive.
	// The goal is to keep track of any references secrets, so these can be redacted easily.
	IsSensitive bool
}

// DeclarationSideEffects is returned by GetDeclarationsSideEffects()
type DeclarationSideEffects struct {
	// CommandHistory is an ordered list of commands: when performed in sequence,
	// will result in a environment that contains the declared env vars
	CommandHistory []Command
	// ResultEnvironment is returned for reference,
	// it will equal the environment after performing the commands
	ResultEnvironment map[string]Variable
}

// EnvironmentSource implementations can return an initial environment
type EnvironmentSource interface {
	GetEnvironment() map[string]Variable
}

// DefaultEnvironmentSource is a default implementation of EnvironmentSource, returns the current environment
type DefaultEnvironmentSource struct{}

// GetEnvironment returns the current process' environment
func (*DefaultEnvironmentSource) GetEnvironment() map[string]Variable {
	processEnvs := os.Environ()
	envs := make(map[string]Variable)

	for _, env := range processEnvs {
		key, value := SplitEnv(env)
		if key == "" {
			continue
		}

		envs[key] = Variable{
			Key:         key,
			Value:       value,
			IsSensitive: false,
		}
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

// GetDeclarationsSideEffects iterates over the list of ordered new declared variables sequentally and returns the needed
// commands (like os.Setenv) to add the variables to the current environment.
// The current process environment is not changed.
// Variable expansion is done also, every new variable can reference the previous and initial environments (via EnvironmentSource)
// The new variables (models.EnvironmentItemModel) can be defined in the envman definition file, or filled in directly.
// If the source of the variables (models.EnvironmentItemModel) is the bitrise.yml workflow,
// they will be in this order: App secrets; App level envs; Workflow level envs; Additional Step info envs; Input envs.
func GetDeclarationsSideEffects(newEnvs []models.EnvironmentItemModel, envSource EnvironmentSource) (DeclarationSideEffects, error) {
	envs := envSource.GetEnvironment()
	commandHistory := make([]Command, len(newEnvs))

	for i, env := range newEnvs {
		command, err := getDeclarationCommand(env, envs)
		if err != nil {
			return DeclarationSideEffects{}, fmt.Errorf("failed to parse new environment variable (%s): %s", env, err)
		}

		commandHistory[i] = command

		switch command.Action {
		case SetAction:
			envs[command.Variable.Key] = command.Variable
		case UnsetAction:
			delete(envs, command.Variable.Key)
		case SkipAction:
		default:
			return DeclarationSideEffects{}, fmt.Errorf("invalid case for environement declaration action: %#v", command)
		}
	}

	return DeclarationSideEffects{
		CommandHistory:    commandHistory,
		ResultEnvironment: envs,
	}, nil
}

// getDeclarationCommand maps a variable to be daclered (env) to an expanded env key and value.
// The current process environment is not changed.
func getDeclarationCommand(env models.EnvironmentItemModel, envs map[string]Variable) (Command, error) {
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

	mappingFuncFactory := func(envs map[string]Variable, containsSensitiveInfo *bool) func(string) string {
		return func(key string) string {
			if _, ok := envs[key]; !ok {
				return ""
			}

			*containsSensitiveInfo = *containsSensitiveInfo || envs[key].IsSensitive
			return envs[key].Value
		}
	}

	containsSensitiveInfo := options.IsSensitive != nil && *options.IsSensitive
	if options.IsExpand != nil && *options.IsExpand {
		envValue = os.Expand(envValue, mappingFuncFactory(envs, &containsSensitiveInfo))
	}

	return Command{
		Action: SetAction,
		Variable: Variable{
			Key:         envKey,
			Value:       envValue,
			IsSensitive: containsSensitiveInfo,
		},
	}, nil
}

// ExecuteCommand sets the current process's envrionment
func ExecuteCommand(command Command) error {
	switch command.Action {
	case SetAction:
		return os.Setenv(command.Variable.Key, command.Variable.Value)
	case UnsetAction:
		return os.Unsetenv(command.Variable.Key)
	case SkipAction:
		return nil
	default:
		return fmt.Errorf("invalid case for environement declaration action: %#v", command)
	}
}
