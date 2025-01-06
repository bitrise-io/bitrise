package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/urfave/cli"
)

// ValidationItemModel ...
type ValidationItemModel struct {
	IsValid  bool     `json:"is_valid" yaml:"is_valid"`
	Error    string   `json:"error,omitempty" yaml:"error,omitempty"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// ValidationModel ...
type ValidationModel struct {
	Config  *ValidationItemModel `json:"config,omitempty" yaml:"config,omitempty"`
	Secrets *ValidationItemModel `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

// ValidateResponseModel ...
type ValidateResponseModel struct {
	Data     *ValidationModel `json:"data,omitempty" yaml:"data,omitempty"`
	Error    string           `json:"error,omitempty" yaml:"error,omitempty"`
	Warnings []string         `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// NewValidationResponse ...
func NewValidationResponse(validation ValidationModel, warnings ...string) ValidateResponseModel {
	return ValidateResponseModel{
		Data:     &validation,
		Warnings: warnings,
	}
}

// NewValidationError ...
func NewValidationError(err string, warnings ...string) ValidateResponseModel {
	return ValidateResponseModel{
		Error:    err,
		Warnings: warnings,
	}
}

// IsValid ...
func (v ValidationModel) IsValid() bool {
	if v.Config == nil && v.Secrets == nil {
		return false
	}

	if v.Config != nil && !v.Config.IsValid {
		return false
	}

	if v.Secrets != nil && !v.Secrets.IsValid {
		return false
	}

	return true
}

// JSON ...
func (v ValidateResponseModel) JSON() string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`"Failed to marshal validation result (%#v), err: %s"`, v, err)
	}
	return string(bytes)
}

func (v ValidateResponseModel) String() string {
	if v.Error != "" {
		msg := fmt.Sprintf("%s: %s", colorstring.Red("Error"), v.Error)
		if len(v.Warnings) > 0 {
			msg += "\nWarning(s):\n"
			for i, warning := range v.Warnings {
				msg += fmt.Sprintf("- %s", warning)
				if i != len(v.Warnings)-1 {
					msg += "\n"
				}
			}
		}
		return msg
	}

	if v.Data != nil {
		msg := v.Data.String()
		if len(v.Warnings) > 0 {
			msg += "\nWarning(s):\n"
			for i, warning := range v.Warnings {
				msg += fmt.Sprintf("- %s", warning)
				if i != len(v.Warnings)-1 {
					msg += "\n"
				}
			}
		}
		return msg
	}

	return ""
}

// String ...
func (v ValidationModel) String() string {
	msg := ""

	if v.Config != nil {
		config := *v.Config
		if config.IsValid {
			msg += fmt.Sprintf("Config is valid: %s", colorstring.Greenf("%v", true))
		} else {
			msg += fmt.Sprintf("Config is valid: %s", colorstring.Redf("%v", false))
			msg += fmt.Sprintf("\nError: %s", colorstring.Red(config.Error))
		}

		if len(config.Warnings) > 0 {
			msg += "\nWarning(s):\n"
			for i, warning := range config.Warnings {
				msg += fmt.Sprintf("- %s", warning)
				if i != len(config.Warnings)-1 {
					msg += "\n"
				}
			}
		}
	}

	if v.Secrets != nil {
		if v.Config != nil {
			msg += "\n"
		}

		secret := *v.Secrets
		if secret.IsValid {
			msg += fmt.Sprintf("Secret is valid: %s", colorstring.Greenf("%v", true))
		} else {
			msg += fmt.Sprintf("Secret is valid: %s", colorstring.Redf("%v", false))
			msg += fmt.Sprintf("\nError: %s", colorstring.Red(secret.Error))
		}
	}

	return msg
}

func validateBitriseYML(bitriseConfigPath string, bitriseConfigBase64Data string) (*ValidationItemModel, error) {
	pth, err := GetBitriseConfigFilePath(bitriseConfigPath)
	if err != nil && !strings.Contains(err.Error(), "bitrise.yml path not defined and not found on it's default path:") {
		return nil, fmt.Errorf("Failed to get config path, err: %s", err)
	}

	if pth != "" || (pth == "" && bitriseConfigBase64Data != "") {
		// Config validation
		_, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath, true)
		configValidation := ValidationItemModel{
			IsValid:  true,
			Warnings: warns,
		}
		if err != nil {
			configValidation.IsValid = false
			configValidation.Error = err.Error()
		}

		return &configValidation, nil
	}

	return nil, nil
}

func validateInventory(inventoryPath string, inventoryBase64Data string) (*ValidationItemModel, error) {
	pth, err := GetInventoryFilePath(inventoryPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get secrets path, err: %s", err)
	}

	if pth != "" || inventoryBase64Data != "" {
		// Inventory validation
		_, err := CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath)
		secretValidation := ValidationItemModel{
			IsValid: true,
		}
		if err != nil {
			secretValidation.IsValid = false
			secretValidation.Error = err.Error()
		}

		return &secretValidation, nil
	}

	return nil, nil
}

func runValidate(bitriseConfigPath string, bitriseConfigBase64Data string, inventoryPath string, inventoryBase64Data string) (*ValidationModel, []string, error) {
	warnings := []string{}

	validation := ValidationModel{}

	result, err := validateBitriseYML(bitriseConfigPath, bitriseConfigBase64Data)
	validation.Config = result
	if err != nil {
		return &validation, warnings, err
	}

	result, err = validateInventory(inventoryPath, inventoryBase64Data)
	validation.Secrets = result
	if err != nil {
		return &validation, warnings, err
	}

	if validation.Config == nil && validation.Secrets == nil {
		return &validation, warnings, fmt.Errorf("No config or secrets found for validation")
	}

	return &validation, warnings, nil
}

func validate(c *cli.Context) error {
	logCommandParameters(c)

	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	bitriseConfigPath := c.String(ConfigKey)

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	format := c.String(OuputFormatKey)
	if format == "" {
		format = output.FormatRaw
	}

	var log Logger
	log = NewDefaultRawLogger()
	if format == output.FormatRaw {
		log = NewDefaultRawLogger()
	} else if format == output.FormatJSON {
		log = NewDefaultJSONLogger()
	} else {
		log.Print(NewValidationError(fmt.Sprintf("Invalid format: %s", format)))
		os.Exit(1)
	}

	validation, warnings, err := runValidate(bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data)
	if err != nil {
		log.Print(NewValidationError(err.Error(), warnings...))
		os.Exit(1)
	}

	log.Print(NewValidationResponse(*validation, warnings...))

	if !validation.IsValid() {
		os.Exit(1)
	}

	return nil
}
