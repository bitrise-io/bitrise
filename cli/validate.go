package cli

import (
	"encoding/json"
	"fmt"

	"os"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	flog "github.com/bitrise-io/go-utils/log"
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

func validate(c *cli.Context) error {
	warnings := []string{}

	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	inventoryBase64Data := c.String(InventoryBase64Key)
	inventoryPath := c.String(InventoryKey)

	format := c.String(OuputFormatKey)
	if format == "" {
		format = output.FormatRaw
	}
	//

	var log flog.Logger
	log = flog.NewDefaultRawLogger()
	if format == output.FormatRaw {
		log = flog.NewDefaultRawLogger()
	} else if format == output.FormatJSON {
		log = flog.NewDefaultJSONLoger()
	} else {
		log.Print(NewValidationError(fmt.Sprintf("Invalid format: %s", format), warnings...))
		os.Exit(1)
	}

	validation := ValidationModel{}

	pth, err := GetBitriseConfigFilePath(bitriseConfigPath)
	if err != nil && err.Error() != "No workflow yml found" {
		log.Print(NewValidationError(fmt.Sprintf("Failed to get config path, err: %s", err), warnings...))
		os.Exit(1)
	}

	if pth != "" || (pth == "" && bitriseConfigBase64Data != "") {
		// Config validation
		_, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
		configValidation := ValidationItemModel{
			IsValid:  true,
			Warnings: warns,
		}
		if err != nil {
			configValidation.IsValid = false
			configValidation.Error = err.Error()
		}

		validation.Config = &configValidation
	}

	pth, err = GetInventoryFilePath(inventoryPath)
	if err != nil {
		log.Print(NewValidationError(fmt.Sprintf("Failed to get secrets path, err: %s", err), warnings...))
		os.Exit(1)
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

		validation.Secrets = &secretValidation
	}

	if validation.Config == nil && validation.Secrets == nil {
		log.Print(NewValidationError("No config or secrets found for validation", warnings...))
		os.Exit(1)
	}

	log.Print(NewValidationResponse(validation, warnings...))

	if !validation.IsValid() {
		os.Exit(1)
	}

	return nil
}
