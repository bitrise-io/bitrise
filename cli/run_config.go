package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/bitrise-io/bitrise/advancedlog"
	"github.com/bitrise-io/bitrise/configs"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
)

func getNoOutputTimoutValue(inventoryEnvironments []envmanModels.EnvironmentItemModel) (string, error) {
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

func readNoOutputTimoutConfiguration(inventoryEnvironments []envmanModels.EnvironmentItemModel) time.Duration {
	const defaultTimeout = -1
	envVal, err := getNoOutputTimoutValue(inventoryEnvironments)
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

func registerNoOutputTimeout(timeout time.Duration) {
	if timeout > 0 {
		msg := fmt.Sprintf("Steps will time out if no output is received for %s.", timeout)
		log.Info(colorstring.Yellow(msg))
	}
	configs.NoOutputTimeout = timeout
}
