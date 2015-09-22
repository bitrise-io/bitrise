package bitrise

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/goinp/goinp"
)

var (
	templateFuncMap = template.FuncMap{
		"getenv": func(key string) string {
			return os.Getenv(key)
		},
		"enveq": func(key, expectedValue string) bool {
			envList, err := EnvmanJSONPrint(InputEnvstorePath)
			if err != nil {
				log.Errorf("Faild to get env list, err: %s", err)
			}
			// outputEnvList, err := EnvmanJSONPrint(OutputEnvstorePath)
			// if err != nil {
			// 	log.Errorf("Faild to get env list, err: %s", err)
			// }
			// for outputKey, outputValue := range outputEnvList {
			// 	envList[outputKey] = outputValue
			// }

			if len(envList) > 0 {
				for aKey, value := range envList {
					if aKey == key {
						return value == expectedValue
					}
				}
			}

			return (os.Getenv(key) == expectedValue)
		},
	}
)

// TemplateDataModel ...
type TemplateDataModel struct {
	BuildResults  models.BuildRunResultsModel
	IsBuildFailed bool
	IsBuildOK     bool
	IsCI          bool
	IsPR          bool
	PullRequestID string
}

func createTemplateDataModel(buildResults models.BuildRunResultsModel) TemplateDataModel {
	pullReqID := os.Getenv(PullRequestIDEnvKey)
	isCI := (os.Getenv(CIModeEnvKey) == "true")
	isBuildOK := !buildResults.IsBuildFailed()
	return TemplateDataModel{
		BuildResults:  buildResults,
		IsBuildFailed: !isBuildOK,
		IsBuildOK:     isBuildOK,
		IsCI:          isCI,
		PullRequestID: pullReqID,
		IsPR:          (pullReqID != ""),
	}
}

// EvaluateStepTemplateToBool ...
func EvaluateStepTemplateToBool(expStr string, buildResults models.BuildRunResultsModel) (bool, error) {
	if expStr == "" {
		return false, errors.New("EvaluateStepTemplateToBool: Invalid, empty input: expStr")
	}

	if !strings.Contains(expStr, "{{") {
		expStr = "{{" + expStr + "}}"
	}

	tmpl := template.New("EvaluateStepTemplateToBool").Funcs(templateFuncMap)
	tmpl, err := tmpl.Parse(expStr)
	if err != nil {
		return false, err
	}

	templateData := createTemplateDataModel(buildResults)
	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, templateData); err != nil {
		return false, err
	}
	resString := resBuffer.String()

	return goinp.ParseBool(resString)
}
