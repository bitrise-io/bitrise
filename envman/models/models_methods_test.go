package models

import (
	"testing"
)

func TestGetKeyValuePair(t *testing.T) {
	t.Log("TestGetKeyValuePair")

    testKey := "test_key"
    testValue := "test_value"
    testTitle := "test_title"
    testDescription:= "test_description"
    testValueOptions := []string{"test_value1", "test_value2"}
    testTrue := true
    testFalse := false

    env := EnvironmentItemModel {
        testKey : testValue,
        OptionsKey: EnvironmentItemOptionsModel {
        Title : "test_title",
        Description: "test_description",
        ValueOptions: []string{"test_value1", "test_value2"},
        IsRequired:
        IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
        IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
    },
    }
}

func TestParseFromInterfaceMap(t *testing.T) {
	t.Log("TestParseFromInterfaceMap")
	// 	t.Fatal("TestParseFromInterfaceMap failed")
}

func TestGetKeyValuePair(t *testing.T) {
	t.Log("TestGetKeyValuePair")
	// 	t.Fatal("TestGetKeyValuePair failed")
}

func TestGetOptions(t *testing.T) {
	t.Log("TestGetOptions")
	// 	t.Fatal("TestGetOptions failed")
}

func TestNormalize(t *testing.T) {
	t.Log("TestNormalize")
	// 	t.Fatal("TestNormalize failed")
}

func TestFillMissingDeafults(t *testing.T) {
	t.Log("TestFillMissingDeafults")
	// 	t.Fatal("TestFillMissingDeafults failed")
}

func TestNormalizeEnvironmentItemModel(t *testing.T) {
	t.Log("TestNormalizeEnvironmentItemModel")
	// 	t.Fatal("TestNormalizeEnvironmentItemModel failed")
}

func TestValidate(t *testing.T) {
	t.Log("TestValidate")
	// 	t.Fatal("TestValidate failed")
}
