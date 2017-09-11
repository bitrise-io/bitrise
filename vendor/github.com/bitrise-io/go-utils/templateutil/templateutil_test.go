package templateutil

import (
	"testing"

	"text/template"

	"github.com/stretchr/testify/require"
)

type EmptyInventory struct {
}

type Inventory struct {
	Material string
	Count    uint
}

func TestEvaluateTemplateStringToString(t *testing.T) {
	t.Log("Empty")
	result, err := EvaluateTemplateStringToString("", EmptyInventory{}, template.FuncMap{})
	require.NoError(t, err)
	require.Equal(t, "", result)
	//
	result, err = EvaluateTemplateStringToString("", Inventory{"wool", 17}, template.FuncMap{})
	require.NoError(t, err)
	require.Equal(t, "", result)
	//
	result, err = EvaluateTemplateStringToString("no template string", Inventory{"wool", 17}, template.FuncMap{})
	require.NoError(t, err)
	require.Equal(t, "no template string", result)

	t.Log("Empty inventory - missing argument/property (error)")
	result, err = EvaluateTemplateStringToString("{{.Count}} items are made of {{.Material}}",
		EmptyInventory{}, template.FuncMap{})
	require.Error(t, err)

	//
	var templateFuncMap = template.FuncMap{
		"isOne": func(i int) bool {
			return i == 1
		},
	}

	t.Log("Simple")
	inv := Inventory{"wool", 17}
	result, err = EvaluateTemplateStringToString("{{.Count}} items are made of {{.Material}}",
		inv, template.FuncMap{})
	require.NoError(t, err)
	require.Equal(t, "17 items are made of wool", result)

	inv = Inventory{"glass", 18}
	result, err = EvaluateTemplateStringToString("{{.Count}} items are made of {{.Material}}",
		inv, templateFuncMap)
	require.NoError(t, err)
	require.Equal(t, "18 items are made of glass", result)
}

func TestEvaluateTemplateStringToStringWithDelimiter(t *testing.T) {
	// custom delimiter
	{
		inv := Inventory{"wool", 17}
		result, err := EvaluateTemplateStringToStringWithDelimiter("<<.Count>> items are made of <<.Material>>",
			inv, template.FuncMap{}, "<<", ">>")
		require.NoError(t, err)
		require.Equal(t, "17 items are made of wool", result)
	}

	// custom delimiter - but defalt used in template
	{
		inv := Inventory{"wool", 17}
		result, err := EvaluateTemplateStringToStringWithDelimiter("{{.Count}} items are made of {{.Material}}",
			inv, template.FuncMap{}, "<<", ">>")
		require.NoError(t, err)
		require.Equal(t, "{{.Count}} items are made of {{.Material}}", result)
	}
}

func Test_evaluateTemplate(t *testing.T) {
	t.Log("No options")
	{
		inv := Inventory{"wool", 17}
		result, err := evaluateTemplate("<<.Count>> items are made of <<.Material>>",
			inv, template.FuncMap{}, "<<", ">>",
			[]string{})
		require.NoError(t, err)
		require.Equal(t, "17 items are made of wool", result)
	}

	t.Log("Template options - error on missing")
	{
		inv := Inventory{"wool", 17}
		result, err := evaluateTemplate("<<.Undefined>> items are made of <<.Material>>",
			inv, template.FuncMap{}, "<<", ">>",
			[]string{"missingkey=error"})
		require.EqualError(t, err, `template: :1:2: executing "" at <.Undefined>: can't evaluate field Undefined in type templateutil.Inventory`)
		require.Equal(t, "", result)
	}

	t.Log("Template options - error on missing; default delimiters")
	{
		inv := Inventory{"wool", 17}
		result, err := evaluateTemplate("{{.UndefinedDefDelim}} items are made of {{.Material}}",
			inv, template.FuncMap{}, "", "",
			[]string{"missingkey=error"})
		require.EqualError(t, err, `template: :1:2: executing "" at <.UndefinedDefDelim>: can't evaluate field UndefinedDefDelim in type templateutil.Inventory`)
		require.Equal(t, "", result)
	}
}

func TestEvaluateTemplateStringToStringWithDelimiterAndOpts(t *testing.T) {
	t.Log("No options")
	{
		inv := Inventory{"wool", 17}
		result, err := EvaluateTemplateStringToStringWithDelimiterAndOpts("<<.Count>> items are made of <<.Material>>",
			inv, template.FuncMap{}, "<<", ">>",
			[]string{})
		require.NoError(t, err)
		require.Equal(t, "17 items are made of wool", result)
	}

	t.Log("Template options - error on missing")
	{
		inv := Inventory{"wool", 17}
		result, err := EvaluateTemplateStringToStringWithDelimiterAndOpts("<<.Undefined>> items are made of <<.Material>>",
			inv, template.FuncMap{}, "<<", ">>",
			[]string{"missingkey=error"})
		require.EqualError(t, err, `template: :1:2: executing "" at <.Undefined>: can't evaluate field Undefined in type templateutil.Inventory`)
		require.Equal(t, "", result)
	}

	t.Log("Template options - error on missing; default delimiters")
	{
		inv := Inventory{"wool", 17}
		result, err := EvaluateTemplateStringToStringWithDelimiterAndOpts("{{.UndefinedDefDelim}} items are made of {{.Material}}",
			inv, template.FuncMap{}, "", "",
			[]string{"missingkey=error"})
		require.EqualError(t, err, `template: :1:2: executing "" at <.UndefinedDefDelim>: can't evaluate field UndefinedDefDelim in type templateutil.Inventory`)
		require.Equal(t, "", result)
	}
}
