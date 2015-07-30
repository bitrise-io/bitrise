package models

import (
	"testing"
)

var (
	testKey          = "test_key"
	testValue        = "test_value"
	testKey1         = "test_key1"
	testValue1       = "test_value1"
	testKey2         = "test_key2"
	testValue2       = "test_value2"
	testTitle        = "test_title"
	testDescription  = "test_description"
	testValueOptions = []string{testKey2, testValue2}
	testTrue         = true
	testFalse        = false
)

func TestGetKeyValuePair(t *testing.T) {
	t.Log("TestGetKeyValuePair")

	// Filled env
	env := EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             &testTitle,
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}

	key, value, err := env.GetKeyValuePair()
	if err != nil {
		t.Fatal(err)
	}

	if key != testKey {
		t.Fatalf("Key (%s) should be: %s", key, testKey)
	}
	if value != testValue {
		t.Fatalf("Value (%s) should be: %s", value, testValue)
	}

	// More then 2 fields
	env = EnvironmentItemModel{
		testKey:  testValue,
		testKey1: testValue1,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             &testTitle,
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("More then 2 fields, should get error")
	}

	// 2 key-value fields
	env = EnvironmentItemModel{
		testKey:  testValue,
		testKey1: testValue1,
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("More then 2 fields, should get error")
	}

	// Not string value
	env = EnvironmentItemModel{
		testKey: true,
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("More then 2 fields, should get error")
	}

	// Empty key
	env = EnvironmentItemModel{
		"": testValue,
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("Empty key, should get error")
	}

	// Missing key-value
	env = EnvironmentItemModel{
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             &testTitle,
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("No key-valu set, should get error")
	}
}

func TestParseFromInterfaceMap(t *testing.T) {
	t.Log("TestParseFromInterfaceMap")

	envOptions := EnvironmentItemOptionsModel{}
	model := map[interface{}]interface{}{}

	// Normal
	model["title"] = testTitle
	model["value_options"] = testValueOptions
	model["is_expand"] = testTrue
	err := envOptions.ParseFromInterfaceMap(model)
	if err != nil {
		t.Fatal(err)
	}

	// Key is not string
	model[true] = "false"
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("Not a string key, should be error")
	}

	// title is not a string
	model = map[interface{}]interface{}{}
	model["title"] = true
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("Title value is not a string, should be error")
	}

	// value_options is not a string slice
	model = map[interface{}]interface{}{}
	model["value_options"] = []interface{}{true, false}
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("value_options is not a string slice, should be error")
	}

	// is_required is not a bool
	model = map[interface{}]interface{}{}
	model["is_required"] = &testTrue
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("is_required is not a bool, should be error")
	}

	// other_key is not supported key
	model = map[interface{}]interface{}{}
	model["other_key"] = testTrue
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("other_key is not a supported key, should be error")
	}
}

func TestGetOptions(t *testing.T) {
	t.Log("TestGetOptions")

	// Filled env
	env := EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:    &testTitle,
			IsExpand: &testFalse,
		},
	}
	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Title == nil || *opts.Title != testTitle {
		t.Fatal("Title is nil, or not correct")
	}
	if opts.IsExpand == nil || *opts.IsExpand != testFalse {
		t.Fatal("IsExpand is nil, or not correct")
	}

	// Missing opts
	env = EnvironmentItemModel{
		testKey: testValue,
	}
	_, err = env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	// Wrong opts
	env = EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: map[string]interface{}{
			"title": &testTitle,
			"test":  &testDescription,
		},
	}
	_, err = env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}
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
