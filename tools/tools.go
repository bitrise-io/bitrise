package tools

import (
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	envman "github.com/bitrise-io/envman/cli"
	envmanEnv "github.com/bitrise-io/envman/env"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
)

// ------------------
// --- Stepman share

// StepmanShare ...
func StepmanShare() error {
	args := []string{"share", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareAudit ...
func StepmanShareAudit() error {
	args := []string{"share", "audit", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareCreate ...
func StepmanShareCreate(tag, git, stepID string) error {
	args := []string{"share", "create", "--tag", tag, "--git", git, "--stepid", stepID, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareFinish ...
func StepmanShareFinish() error {
	args := []string{"share", "finish", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// StepmanShareStart ...
func StepmanShareStart(collection string) error {
	args := []string{"share", "start", "--collection", collection, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// EnvmanInit ...
func EnvmanInit(envStorePth string, clear bool) error {
	return envman.InitEnvStore(envStorePth, clear)
}

// EnvmanAdd ...
func EnvmanAdd(envStorePth, key, value string, expand, skipIfEmpty, sensitive bool) error {
	return envman.AddEnv(envStorePth, key, value, expand, false, skipIfEmpty, sensitive)
}

// EnvmanAddEnvs ...
func EnvmanAddEnvs(envstorePth string, envsList []envmanModels.EnvironmentItemModel) error {
	for _, env := range envsList {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return err
		}

		isExpand := envmanModels.DefaultIsExpand
		if opts.IsExpand != nil {
			isExpand = *opts.IsExpand
		}

		skipIfEmpty := envmanModels.DefaultSkipIfEmpty
		if opts.SkipIfEmpty != nil {
			skipIfEmpty = *opts.SkipIfEmpty
		}

		sensitive := envmanModels.DefaultIsSensitive
		if opts.IsSensitive != nil {
			sensitive = *opts.IsSensitive
		}

		if err := EnvmanAdd(envstorePth, key, value, isExpand, skipIfEmpty, sensitive); err != nil {
			return err
		}
	}
	return nil
}

// EnvmanReadEnvList ...
func EnvmanReadEnvList(envStorePth string) (envmanModels.EnvsJSONListModel, error) {
	return envman.ReadEnvsJSONList(envStorePth, true, false, &envmanEnv.DefaultEnvironmentSource{})
}

// EnvmanClear ...
func EnvmanClear(envStorePth string) error {
	return envman.ClearEnvs(envStorePth)
}

// ------------------
// --- Utility

// GetSecretKeysAndValues filters out built in configuration parameters from the secret envs
func GetSecretKeysAndValues(secrets []envmanModels.EnvironmentItemModel) ([]string, []string) {
	var secretKeys []string
	var secretValues []string
	for _, secret := range secrets {
		key, value, err := secret.GetKeyValuePair()
		if err != nil || len(value) < 1 || IsBuiltInFlagTypeKey(key) {
			if err != nil {
				log.Warnf("Error getting key-value pair from secret (%v): %s", secret, err)
			}
			continue
		}
		secretKeys = append(secretKeys, key)
		secretValues = append(secretValues, value)
	}

	return secretKeys, secretValues
}

// IsBuiltInFlagTypeKey returns true if the env key is a built-in flag type env key
func IsBuiltInFlagTypeKey(env string) bool {
	switch env {
	case configs.IsSecretFilteringKey,
		configs.IsSecretEnvsFilteringKey,
		configs.CIModeEnvKey,
		configs.PRModeEnvKey,
		configs.DebugModeEnvKey,
		configs.PullRequestIDEnvKey:
		return true
	default:
		return false
	}
}
