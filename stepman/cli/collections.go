package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/go-utils/colorstring"
	flog "github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

// -------------
// Output Models

// OutputModel ...
type OutputModel struct {
	Data  *([]models.SteplibInfoModel) `json:"data,omitempty" yaml:"data,omitempty"`
	Error string                       `json:"error,omitempty" yaml:"error,omitempty"`
}

// String ...
func (output OutputModel) String() string {
	if output.Error != "" {
		return fmt.Sprintf("%s: %s", colorstring.Red("Error"), output.Error)
	}

	if output.Data == nil {
		return ""
	}

	str := ""
	steplibInfos := *output.Data
	for idx, steplibInfo := range steplibInfos {
		str += colorstring.Bluef("%s\n", steplibInfo.URI)
		str += fmt.Sprintf("  spec_path: %s\n", steplibInfo.SpecPath)
		if idx != len(steplibInfos)-1 {
			str += "\n"
		}
	}
	return str
}

// JSON ...
func (output OutputModel) JSON() string {
	bytes, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf(`"Failed to marshal output (%#v), err: %s"`, output, err)
	}
	return string(bytes)
}

// NewOutput ...
func NewOutput(steplibInfos []models.SteplibInfoModel) OutputModel {
	return OutputModel{
		Data: &steplibInfos,
	}
}

// NewErrorOutput ...
func NewErrorOutput(err error) OutputModel {
	return OutputModel{
		Error: err.Error(),
	}
}

// -------------

func collections(c *cli.Context) error {
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	}

	var log flog.Logger
	if format == OutputFormatRaw {
		log = flog.NewDefaultRawLogger()
	} else if format == OutputFormatJSON {
		log = flog.NewDefaultJSONLoger()
	} else {
		failf("invalid format: %s", format)
	}

	steplibInfos, err := Collections()
	if err != nil {
		out := NewErrorOutput(err)
		if format == OutputFormatJSON {
			failf(out.JSON())
		}
		failf(out.String())
	}

	log.Print(NewOutput(steplibInfos))

	return nil
}

// Collections returns SteplibInfoModels about the locally configured step collections.
func Collections() ([]models.SteplibInfoModel, error) {
	var steplibInfos []models.SteplibInfoModel

	stepLibURIs := stepman.GetAllStepCollectionPath()
	for _, steplibURI := range stepLibURIs {
		route, found := stepman.ReadRoute(steplibURI)
		if !found {
			return nil, fmt.Errorf("no routing found for steplib: %s", steplibURI)
		}

		specPth := stepman.GetStepSpecPath(route)

		steplibInfos = append(steplibInfos, models.SteplibInfoModel{
			URI:      steplibURI,
			SpecPath: specPth,
		})
	}

	return steplibInfos, nil
}
