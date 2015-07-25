package bitrise

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"text/template"

	"github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/goinp/goinp"
)

var (
	templateFuncMap = template.FuncMap{
		"getenv": func(key string) string {
			return os.Getenv(key)
		},
		"enveq": func(key, expectedValue string) bool {
			return (os.Getenv(key) == expectedValue)
		},
	}
)

// TemplateDataModel ...
type TemplateDataModel struct {
	BuildResults  models.StepRunResultsModel
	IsBuildFailed bool
	IsCI          bool
}

func createTemplateDataModel(buildResults models.StepRunResultsModel, isCI bool) TemplateDataModel {
	return TemplateDataModel{
		BuildResults:  buildResults,
		IsBuildFailed: buildResults.IsBuildFailed(),
		IsCI:          isCI,
	}
}

// EvaluateStepTemplateToBool ...
func EvaluateStepTemplateToBool(expStr string, buildResults models.StepRunResultsModel, isCI bool) (bool, error) {
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

	templateData := createTemplateDataModel(buildResults, isCI)
	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, templateData); err != nil {
		return false, err
	}
	resString := resBuffer.String()

	return goinp.ParseBool(resString)
}
