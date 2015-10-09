package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

// ValidationItemModel ...
type ValidationItemModel struct {
	IsValid bool   `json:"is_valid" yaml:"is_valid"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ValidationModel ...
type ValidationModel struct {
	Config  *ValidationItemModel `json:"config,omitempty" yaml:"config,omitempty"`
	Secrets *ValidationItemModel `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

func printRawValidation(validation ValidationModel) error {
	validConfig := true
	if validation.Config != nil {
		fmt.Println(colorstring.Blue("Config validation result:"))
		configValidation := *validation.Config
		if configValidation.IsValid {
			fmt.Printf("is valid: %s\n", colorstring.Greenf("%v", configValidation.IsValid))
		} else {
			fmt.Printf("is valid: %s\n", colorstring.Redf("%v", configValidation.IsValid))
			fmt.Printf("error: %s\n", colorstring.Red(configValidation.Error))

			validConfig = false
		}
		fmt.Println()
	}

	validSecrets := true
	if validation.Secrets != nil {
		fmt.Println(colorstring.Blue("Secret validation result:"))
		secretValidation := *validation.Secrets
		if secretValidation.IsValid {
			fmt.Printf("is valid: %s\n", colorstring.Greenf("%v", secretValidation.IsValid))
		} else {
			fmt.Printf("is valid: %s\n", colorstring.Redf("%v", secretValidation.IsValid))
			fmt.Printf("error: %s\n", colorstring.Red(secretValidation.Error))

			validSecrets = false
		}
	}

	if !validConfig && !validSecrets {
		return errors.New("Config and secrets are invalid")
	} else if !validConfig {
		return errors.New("Config is invalid")
	} else if !validSecrets {
		return errors.New("Secret is invalid")
	}
	return nil
}

func printJSONValidation(validation ValidationModel) {
	bytes, err := json.Marshal(validation)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to parse validation result, err: %s, result: %#v", err, validation), OutputFormatJSON)
	}

	fmt.Println(string(bytes))
}

func validate(c *cli.Context) {
	format := c.String(OuputFormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	validation := ValidationModel{}

	pth, err := GetBitriseConfigFilePath(c)
	if err != nil && err.Error() != "No workflow yml found" {
		registerFatal(fmt.Sprintf("Failed to get config path, err: %s", err), format)
	}
	if pth != "" || (pth == "" && c.String(ConfigBase64Key) != "") {
		// Config validation
		isValid := true
		errMsg := ""

		_, err := CreateBitriseConfigFromCLIParams(c)
		if err != nil {
			isValid = false
			errMsg = err.Error()
		}

		validation.Config = &ValidationItemModel{
			IsValid: isValid,
			Error:   errMsg,
		}
	} else {
		log.Debug("No config found for validation")
	}

	pth, err = GetInventoryFilePath(c)
	if err != nil {
		registerFatal(fmt.Sprintf("Failed to get secrets path, err: %s", err), format)
	}
	if pth != "" || c.String(InventoryBase64Key) != "" {
		// Inventory validation
		isValid := true
		errMsg := ""

		_, err := CreateInventoryFromCLIParams(c)
		if err != nil {
			isValid = false
			errMsg = err.Error()
		}

		validation.Secrets = &ValidationItemModel{
			IsValid: isValid,
			Error:   errMsg,
		}
	}

	if validation.Config == nil && validation.Secrets == nil {
		registerFatal("No config or secrets found for validation", format)
	}

	switch format {
	case OutputFormatRaw:
		if err := printRawValidation(validation); err != nil {
			registerFatal(fmt.Sprintf("Validation failed, err: %s", err), format)
		}
		break
	case OutputFormatJSON:
		printJSONValidation(validation)
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}
}
