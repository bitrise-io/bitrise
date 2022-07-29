package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bitrise-io/bitrise/configs"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	log "github.com/sirupsen/logrus"
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
	const defaultTimeout = 0
	envVal, err := getNoOutputTimoutValue(inventoryEnvironments)
	if err != nil {
		log.Errorf("Failed to read value of %s: %s", configs.NoOutputTimeoutEnvKey, err)
		return defaultTimeout
	}

	if envVal == "" {
		return defaultTimeout
	}

	timeout, err := strconv.ParseUint(envVal, 10, 0)
	if err != nil {
		log.Errorf("Invalid configuration environment variable value $%s=%s: %s", configs.NoOutputTimeoutEnvKey, envVal, err)
		return defaultTimeout
	}

	return time.Duration(timeout) * time.Second
}

func registerNoOutputTimeout(timeout time.Duration) {
	if timeout != 0 {
		msg := fmt.Sprintf("Steps are being aborted after not receiving output for %s.", timeout)
		log.Info(colorstring.Yellow(msg))
	}
	configs.NoOutputTimeout = timeout
}
