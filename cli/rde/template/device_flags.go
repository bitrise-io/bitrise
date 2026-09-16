package template

import (
	"fmt"

	"github.com/spf13/cobra"

	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

// deviceFlags collects the --device-* flags 'template create' and
// 'template update' share. They mirror the flags of 'session create' so a
// device declared on a template reads the same as one requested per session.
type deviceFlags struct {
	platform    string
	model       string
	osVersion   string
	systemImage string
}

func (d *deviceFlags) bind(c *cobra.Command) {
	c.Flags().StringVar(&d.platform, "device-platform", "", "declare a virtual device sessions created from the template boot unless overridden: ios (simulator, macOS stack) or android (emulator, Linux stack); read 'rde device-guide' first")
	c.Flags().StringVar(&d.model, "device-model", "", "device to boot: simctl device type (\"iPhone 16\") or emulator device profile (\"pixel_7\"); default: platform default; requires --device-platform")
	c.Flags().StringVar(&d.osVersion, "device-os-version", "", "iOS only: an iOS version (\"18.2\") or simctl runtime id — anything else is rejected; default: newest installed; requires --device-platform")
	c.Flags().StringVar(&d.systemImage, "device-system-image", "", "Android only: sdkmanager system image package (\"system-images;android-34;google_apis;x86_64\"); default: platform default; requires --device-platform")
}

// set reports whether any device flag was given.
func (d *deviceFlags) set() bool {
	return d.platform != "" || d.model != "" || d.osVersion != "" || d.systemImage != ""
}

// spec validates the flags and returns the device they declare, or nil when
// none was given. Declaring a device on a template always needs the
// platform: there is nothing to inherit it from.
func (d *deviceFlags) spec() (*internalrde.DeviceSpec, error) {
	if !d.set() {
		return nil, nil
	}
	switch d.platform {
	case "ios", "android":
	case "":
		return nil, fmt.Errorf("--device-model, --device-os-version and --device-system-image require --device-platform")
	default:
		return nil, fmt.Errorf("--device-platform must be ios or android")
	}
	if d.platform == "ios" && d.systemImage != "" {
		return nil, fmt.Errorf("--device-system-image applies to Android only")
	}
	return &internalrde.DeviceSpec{
		Platform:    d.platform,
		DeviceModel: d.model,
		OSVersion:   d.osVersion,
		SystemImage: d.systemImage,
	}, nil
}
