package yml

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/configmerge"
	internalyml "github.com/bitrise-io/bitrise/v2/internal/yml"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/spf13/cobra"
)

// offlineKey skips online validation even when authenticated, forcing the
// historical local-only behavior deterministically.
const offlineKey = "offline"

// NewValidateCommand ...
func NewValidateCommand() *cobra.Command {
	validateCommand := &cobra.Command{
		Use:   "validate",
		Short: "Validates a specified bitrise config.",
		RunE:  validate,
	}

	cmdutil.AddConfigAndInventoryFlags(validateCommand.Flags())
	validateCommand.Flags().String(cmdutil.FormatKey, "", "Output format. Accepted: raw (default), json.")
	validateCommand.Flags().Bool(offlineKey, false, "Skip online validation even if authenticated; use only the local schema check.")
	cmdutil.AddAppFlag(validateCommand.Flags(), "app ID to validate against (enables app-specific checks: stacks, machine types, license pools; inside a build, defaults to the app the build runs for)")

	// --offline skips the online path entirely, so accepting it together with
	// --app would silently ignore the app-specific checks --app asks for. An
	// ambient BITRISE_APP_ID/BITRISE_APP_SLUG isn't covered (cobra only sees
	// flags) and stays silently ignored under --offline, since neither is an
	// explicit request.
	validateCommand.MarkFlagsMutuallyExclusive(offlineKey, cmdutil.FlagApp)

	return validateCommand
}

// sourceOnline marks a result produced by the API instead of the local
// schema check. Local results leave Source empty rather than naming
// themselves, so --offline output stays byte-identical to v2.
const sourceOnline = "online"

// ValidationItemModel ...
type ValidationItemModel struct {
	IsValid  bool     `json:"is_valid" yaml:"is_valid"`
	Error    string   `json:"error,omitempty" yaml:"error,omitempty"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Source   string   `json:"source,omitempty" yaml:"source,omitempty"`
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
		msg += formatWarnings(v.Warnings)
		return msg
	}

	if v.Data != nil {
		msg := v.Data.String()
		msg += formatWarnings(v.Warnings)
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

		msg += formatWarnings(config.Warnings)
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

func formatWarnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}
	msg := "\nWarning(s):\n"
	for i, warning := range warnings {
		msg += fmt.Sprintf("- %s", warning)
		if i != len(warnings)-1 {
			msg += "\n"
		}
	}
	return msg
}

// hasConfigToValidate reports whether a config was actually specified via
// -c/--config-base64 (or resolves via the default ./bitrise.yml path), so
// callers can treat "nothing to validate" (e.g. running with -i only) as a
// no-op rather than a path-resolution failure.
func hasConfigToValidate(bitriseConfigPath, bitriseConfigBase64Data string) (bool, error) {
	pth, err := cmdutil.GetBitriseConfigFilePath(bitriseConfigPath)
	if err != nil && !errors.Is(err, cmdutil.ErrNoBitriseConfigFound) {
		return false, fmt.Errorf("failed to get config path, err: %s", err)
	}
	return pth != "" || bitriseConfigBase64Data != "", nil
}

// validateBitriseYML runs the local schema check. Callers gate on
// hasConfigToValidate first, so there is always a config to validate here.
func validateBitriseYML(bitriseConfigPath string, bitriseConfigBase64Data string) *ValidationItemModel {
	_, warns, err := cmdutil.CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath, bitrise.ValidationTypeFull)
	configValidation := ValidationItemModel{
		IsValid:  true,
		Warnings: warns,
	}
	if err != nil {
		configValidation.IsValid = false
		configValidation.Error = err.Error()
	}

	return &configValidation
}

func validateInventory(inventoryPath string, inventoryBase64Data string) (*ValidationItemModel, error) {
	pth, err := cmdutil.GetInventoryFilePath(inventoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets path, err: %s", err)
	}

	if pth != "" || inventoryBase64Data != "" {
		// Inventory validation
		_, err := cmdutil.CreateInventoryFromCLIParams(inventoryBase64Data, inventoryPath)
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

// getYmlStringForOnlineValidation returns the flattened YAML text to submit for
// online validation: the file/base64 content as-is for a plain config, or
// the merged result for a modular one
func getYmlStringForOnlineValidation(bitriseConfigPath, bitriseConfigBase64Data string) (string, error) {
	if bitriseConfigBase64Data != "" {
		data, err := base64.StdEncoding.DecodeString(bitriseConfigBase64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base 64 string, error: %s", err)
		}
		return string(data), nil
	}

	pth, err := cmdutil.GetBitriseConfigFilePath(bitriseConfigPath)
	if err != nil {
		return "", err
	}

	isModularConfig, err := configmerge.IsModularConfig(pth)
	if err != nil {
		return "", fmt.Errorf("failed to check if the config is modular: %s", err)
	}
	if !isModularConfig {
		content, err := fileutil.ReadBytesFromFile(pth)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	merger, err := cmdutil.CreateDefaultMerger()
	if err != nil {
		return "", fmt.Errorf("failed to create config module merger: %w", err)
	}
	mergedConfigContent, _, err := merger.MergeConfig(pth)
	if err != nil {
		return "", fmt.Errorf("failed to merge Bitrise config (%s): %w", pth, err)
	}
	return mergedConfigContent, nil
}

// tryOnlineValidate attempts online validation, returning ok=true only when
// the attempt completed — regardless of the config turning out valid or
// not, that's a complete result on its own (the API performs the same
// schema checks as the local one, plus app-specific ones), so the caller
// should not also run local validation in that case. ok=false with an empty
// warning means there's simply no token (the expected default, not a
// degradation); ok=false with a warning means the attempt itself couldn't
// be completed (preparing the submit text, network, 5xx, timeout), and the
// caller should fall back to local validation.
func tryOnlineValidate(cmd *cobra.Command, bitriseConfigPath, bitriseConfigBase64Data, appSlug string) (item *ValidationItemModel, warning string, ok bool) {
	client, err := cmdutil.NewAPIClient(cmd)
	if err != nil {
		if errors.Is(err, cmdutil.ErrNoToken) {
			return nil, "", false
		}
		return nil, fmt.Sprintf("online validation unavailable: %s", err), false
	}

	rawYAML, err := getYmlStringForOnlineValidation(bitriseConfigPath, bitriseConfigBase64Data)
	if err != nil {
		return nil, fmt.Sprintf("online validation unavailable: %s", err), false
	}

	result, err := internalyml.NewService(client).Validate(cmd.Context(), rawYAML, appSlug)
	if err != nil {
		return nil, fmt.Sprintf("online validation unavailable: %s", err), false
	}

	item = &ValidationItemModel{IsValid: result.Valid, Warnings: result.Warnings, Source: sourceOnline}
	if len(result.Errors) > 0 {
		item.Error = strings.Join(result.Errors, "; ")
	}
	return item, "", true
}

// validateConfig prefers the online validate-bitrise-yml endpoint when a
// token is available and --offline wasn't passed. Local validation only
// runs when there's no token, --offline was passed, or the online attempt
// itself couldn't be completed — in which case the returned warning
// explains why, and the local result is used instead.
func validateConfig(cmd *cobra.Command, bitriseConfigPath, bitriseConfigBase64Data string, offline bool, appSlug string) (*ValidationItemModel, string, error) {
	hasConfig, err := hasConfigToValidate(bitriseConfigPath, bitriseConfigBase64Data)
	if err != nil {
		return nil, "", err
	}
	if !hasConfig {
		return nil, "", nil
	}

	if !offline {
		if item, warning, ok := tryOnlineValidate(cmd, bitriseConfigPath, bitriseConfigBase64Data, appSlug); ok {
			return item, "", nil
		} else if warning != "" {
			return validateBitriseYML(bitriseConfigPath, bitriseConfigBase64Data), warning, nil
		}
	}

	return validateBitriseYML(bitriseConfigPath, bitriseConfigBase64Data), "", nil
}

func runValidate(cmd *cobra.Command, bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data string, offline bool, appSlug string) (*ValidationModel, []string, error) {
	warnings := []string{}

	validation := ValidationModel{}

	configItem, warning, err := validateConfig(cmd, bitriseConfigPath, bitriseConfigBase64Data, offline, appSlug)
	validation.Config = configItem
	if warning != "" {
		warnings = append(warnings, warning)
	}
	if err != nil {
		return &validation, warnings, err
	}

	result, err := validateInventory(inventoryPath, inventoryBase64Data)
	validation.Secrets = result
	if err != nil {
		return &validation, warnings, err
	}

	if validation.Config == nil && validation.Secrets == nil {
		return &validation, warnings, fmt.Errorf("no config or secrets found for validation")
	}

	return &validation, warnings, nil
}

func validate(cmd *cobra.Command, _ []string) error {
	cmdutil.LogCommandParameters(cmd)

	bitriseConfigBase64Data, _ := cmd.Flags().GetString(cmdutil.ConfigBase64Key)
	bitriseConfigPath, _ := cmd.Flags().GetString(cmdutil.ConfigKey)

	inventoryBase64Data, _ := cmd.Flags().GetString(cmdutil.InventoryBase64Key)
	inventoryPath, _ := cmd.Flags().GetString(cmdutil.InventoryKey)

	offline, _ := cmd.Flags().GetBool(offlineKey)
	appSlug := cmdutil.LookupAppSlug(cmd)

	format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
	if format == "" {
		format = output.FormatRaw
	}

	var log cmdutil.Logger = cmdutil.NewDefaultRawLogger()
	if format == output.FormatJSON {
		log = cmdutil.NewDefaultJSONLogger()
	} else if format != output.FormatRaw {
		log.Print(NewValidationError(fmt.Sprintf("Invalid format: %s", format)))
		os.Exit(1)
	}

	validation, warnings, err := runValidate(cmd, bitriseConfigPath, bitriseConfigBase64Data, inventoryPath, inventoryBase64Data, offline, appSlug)

	// The ambient-token default means `validate` can validate online without
	// the user asking for it, so say so — on stderr, to keep the result on
	// stdout machine-readable.
	if validation != nil && validation.Config != nil && validation.Config.Source == sourceOnline {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Validated online. Use --%s to run the local check instead.\n", offlineKey)
	}

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
