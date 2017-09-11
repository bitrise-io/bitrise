package templateutil

import (
	"bytes"
	"text/template"
)

// evaluateTemplate ...
//
// templateOptions: https://golang.org/pkg/text/template/#Template.Option
func evaluateTemplate(
	templateContent string,
	inventory interface{},
	funcs template.FuncMap,
	delimLeft, delimRight string,
	templateOptions []string,
) (string, error) {
	tmpl := template.New("").Funcs(funcs).Delims(delimLeft, delimRight)
	if len(templateOptions) > 0 {
		tmpl = tmpl.Option(templateOptions...)
	}
	tmpl, err := tmpl.Parse(templateContent)
	if err != nil {
		return "", err
	}

	var resBuffer bytes.Buffer
	if err := tmpl.Execute(&resBuffer, inventory); err != nil {
		return "", err
	}

	return resBuffer.String(), nil
}

// EvaluateTemplateStringToStringWithDelimiterAndOpts ...
//
// templateOptions: https://golang.org/pkg/text/template/#Template.Option
func EvaluateTemplateStringToStringWithDelimiterAndOpts(
	templateContent string,
	inventory interface{},
	funcs template.FuncMap,
	delimLeft, delimRight string,
	templateOptions []string,
) (string, error) {
	return evaluateTemplate(templateContent, inventory, funcs, delimLeft, delimRight, templateOptions)
}

// EvaluateTemplateStringToStringWithDelimiter ...
func EvaluateTemplateStringToStringWithDelimiter(
	templateContent string,
	inventory interface{},
	funcs template.FuncMap,
	delimLeft, delimRight string,
) (string, error) {
	return evaluateTemplate(templateContent, inventory, funcs, delimLeft, delimRight, []string{})
}

// EvaluateTemplateStringToString ...
func EvaluateTemplateStringToString(templateContent string, inventory interface{}, funcs template.FuncMap) (string, error) {
	return EvaluateTemplateStringToStringWithDelimiter(templateContent, inventory, funcs, "", "")
}
