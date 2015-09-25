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

// TemplateDataModel ...
type TemplateDataModel struct {
	BuildResults  models.BuildRunResultsModel
	IsBuildFailed bool
	IsBuildOK     bool
	IsCI          bool
	IsPR          bool
	PullRequestID string
}

var (
	templateFuncMap = template.FuncMap{
		"getenv": func(key string) string {
			return getEnv(key)
		},
		"enveq": func(key, expectedValue string) bool {
			return (getEnv(key) == expectedValue)
		},
	}
)

func getEnv(key string) string {
	envList, err := EnvmanJSONPrint(InputEnvstorePath)
	if err != nil && !strings.Contains(err.Error(), "Faild to read envs, error: No file found at path") {
		log.Errorf("Faild to get env list, err: %s", err)
	}
	if len(envList) > 0 {
		for aKey, value := range envList {
			if aKey == key {
				return value
			}
		}
	}
	return os.Getenv(key)
}

func createTemplateDataModel(isCI bool, pullReqID string, buildResults models.BuildRunResultsModel) TemplateDataModel {
	isBuildOK := !buildResults.IsBuildFailed()
	IsPullRequestMode := (pullReqID != "")

	return TemplateDataModel{
		BuildResults:  buildResults,
		IsBuildFailed: !isBuildOK,
		IsBuildOK:     isBuildOK,
		IsCI:          isCI,
		PullRequestID: pullReqID,
		IsPR:          IsPullRequestMode,
	}
}

// EvaluateTemplateToBool ...
func EvaluateTemplateToBool(expStr string, isCI bool, pullReqID string, buildResults models.BuildRunResultsModel) (bool, error) {
	if expStr == "" {
		return false, errors.New("EvaluateTemplateToBool: Invalid, empty input: expStr")
	}

	if !strings.Contains(expStr, "{{") {
		expStr = "{{" + expStr + "}}"
	}

	tmpl := template.New("EvaluateTemplateToBool").Funcs(templateFuncMap)
	tmpl, err := tmpl.Parse(expStr)
	if err != nil {
		return false, err
	}

	templateData := createTemplateDataModel(isCI, pullReqID, buildResults)
	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, templateData); err != nil {
		return false, err
	}
	resString := resBuffer.String()

	return goinp.ParseBool(resString)
}

// EvaluateTemplateToString ...
func EvaluateTemplateToString(expStr string, isCI bool, pullReqID string, buildResults models.BuildRunResultsModel) (string, error) {
	if expStr == "" {
		return "", errors.New("EvaluateTemplateToBool: Invalid, empty input: expStr")
	}

	if !strings.Contains(expStr, "{{") {
		expStr = "{{" + expStr + "}}"
	}

	tmpl := template.New("EvaluateTemplateToBool").Funcs(templateFuncMap)
	tmpl, err := tmpl.Parse(expStr)
	if err != nil {
		return "", err
	}

	templateData := createTemplateDataModel(isCI, pullReqID, buildResults)
	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, templateData); err != nil {
		return "", err
	}

	return resBuffer.String(), nil
}
