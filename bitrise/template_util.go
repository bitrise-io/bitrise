package bitrise

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"text/template"

	"github.com/bitrise-io/bitrise/v2/envfile"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/goinp/goinp"
)

type TemplateDataModel struct {
	BuildResults  models.BuildRunResultsModel
	IsBuildFailed bool
	IsBuildOK     bool
	IsCI          bool
	IsPR          bool
	PullRequestID string
}

func getEnv(key string, envList envmanModels.EnvsJSONListModel) string {
	envfilePath := os.Getenv(envfile.DefaultEnvfilePathEnv)
	_, err := os.Stat(envfilePath)
	if errors.Is(err, os.ErrNotExist) {
		for aKey, value := range envList {
			if aKey == key {
				return value
			}
		}
		return os.Getenv(key)
	}

	value, err := envfile.GetEnv(key, envList, envfilePath)
	if err != nil {
		log.Warnf("Failed to read env from envfile: %s", err)
		log.Warnf("Falling back to the current value of $%s", key)
		return os.Getenv(key)
	}

	return value
}

func createTemplateDataModel(isCI, isPR bool, buildResults models.BuildRunResultsModel) TemplateDataModel {
	isBuildOK := !buildResults.IsBuildFailed()

	return TemplateDataModel{
		BuildResults:  buildResults,
		IsBuildFailed: !isBuildOK,
		IsBuildOK:     isBuildOK,
		IsCI:          isCI,
		IsPR:          isPR,
	}
}

func EvaluateTemplateToString(expStr string, isCI, isPR bool, buildResults models.BuildRunResultsModel, envList envmanModels.EnvsJSONListModel) (string, error) {
	if expStr == "" {
		return "", errors.New("EvaluateTemplateToBool: Invalid, empty input: expStr")
	}

	if !strings.Contains(expStr, "{{") {
		expStr = "{{" + expStr + "}}"
	}

	var templateFuncMap = template.FuncMap{
		"getenv": func(key string) string {
			return getEnv(key, envList)
		},
		"enveq": func(key, expectedValue string) bool {
			return (getEnv(key, envList) == expectedValue)
		},
		"envcontain": func(key, subString string) bool {
			return strings.Contains(getEnv(key, envList), subString)
		},
	}

	tmpl := template.New("EvaluateTemplateToBool").Funcs(templateFuncMap)
	tmpl, err := tmpl.Parse(expStr)
	if err != nil {
		return "", err
	}

	templateData := createTemplateDataModel(isCI, isPR, buildResults)
	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, templateData); err != nil {
		return "", err
	}

	return resBuffer.String(), nil
}

func EvaluateTemplateToBool(expStr string, isCI, isPR bool, buildResults models.BuildRunResultsModel, envList envmanModels.EnvsJSONListModel) (bool, error) {
	resString, err := EvaluateTemplateToString(expStr, isCI, isPR, buildResults, envList)
	if err != nil {
		return false, err
	}

	return goinp.ParseBool(resString)
}
