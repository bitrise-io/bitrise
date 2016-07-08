package cli

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func registerFatal(errorMsg, format string) {
	msg := map[string]string{
		"error": errorMsg,
	}

	if format == OutputFormatRaw {
		log.Fatal(msg["error"])
	} else {
		bytes, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Failed to parse error model, err: %s", err)
		}

		fmt.Println(string(bytes))
		os.Exit(1)
	}
}

func collections(c *cli.Context) error {
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	stepLibURIs := stepman.GetAllStepCollectionPath()

	switch format {
	case OutputFormatRaw:
		for _, stepLibURI := range stepLibURIs {
			fmt.Println(colorstring.Blue(stepLibURI))
		}
		break
	case OutputFormatJSON:
		stepLibs := models.StepLibURIsModel{}
		for _, stepLibURI := range stepLibURIs {
			stepLibs.StepLibURIs = append(stepLibs.StepLibURIs, models.StepLibURIModel{URI: stepLibURI})
		}
		bytes, err := json.Marshal(stepLibs)
		if err != nil {
			registerFatal(err.Error(), format)
		}
		fmt.Println(string(bytes))
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	return nil
}
