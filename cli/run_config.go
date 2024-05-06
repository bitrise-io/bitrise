package cli

import (
	"os"
	"strconv"
	"time"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	envmanModels "github.com/bitrise-io/envman/models"
)

func getNoOutputTimeoutValue(inventoryEnvironments []envmanModels.EnvironmentItemModel) (string, error) {
	for _, env := range inventoryEnvironments {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return "", err
		}

		if key == configs.NoOutputTimeoutEnvKey && value != "" {
			return value, nil
		}
	}

	return os.Getenv(configs.NoOutputTimeoutEnvKey), nil
}

func readNoOutputTimeoutConfiguration(inventoryEnvironments []envmanModels.EnvironmentItemModel) time.Duration {
	const defaultTimeout = -1
	envVal, err := getNoOutputTimeoutValue(inventoryEnvironments)
	if err != nil {
		log.Errorf("Failed to read value of %s: %s", configs.NoOutputTimeoutEnvKey, err)
		return defaultTimeout
	}

	if envVal == "" {
		return defaultTimeout
	}

	timeout, err := strconv.ParseInt(envVal, 10, 0)
	if err != nil {
		log.Errorf("Invalid configuration environment variable value $%s=%s: %s", configs.NoOutputTimeoutEnvKey, envVal, err)
		return defaultTimeout
	}

	if timeout <= 0 {
		timeout = -1
	}

	return time.Duration(timeout) * time.Second
}
