package bitrise

import (
	"bytes"
	"os"
	"text/template"

	"github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
	"github.com/bitrise-io/goinp/goinp"
)

var (
	funcMap = template.FuncMap{
		"getenv": func(key string) string { return os.Getenv(key) },
	}
)

// EvaluateStepTemplateToBool ...
func EvaluateStepTemplateToBool(expStr string, buildResults models.StepRunResultsModel) (bool, error) {
	tmpl := template.New("EvaluateStepTemplateToBool").Funcs(funcMap)
	tmpl, err := tmpl.Parse(expStr)
	if err != nil {
		return false, err
	}

	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, buildResults); err != nil {
		return false, err
	}
	resString := resBuffer.String()

	return goinp.ParseBool(resString)
}
