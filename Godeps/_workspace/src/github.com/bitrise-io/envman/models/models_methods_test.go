package models

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
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
	testSummary      = "test_summary"
	testValueOptions = []string{testKey2, testValue2}
	testTrue         = true
	testFalse        = false
)

func TestGetKeyValuePair(t *testing.T) {
	// Filled env
	env := EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			Summary:           pointers.NewStringPtr(testSummary),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
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
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
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
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
		},
	}

	key, value, err = env.GetKeyValuePair()
	if err == nil {
		t.Fatal("No key-valu set, should get error")
	}
}

func TestParseFromInterfaceMap(t *testing.T) {
	envOptions := EnvironmentItemOptionsModel{}
	model := map[string]interface{}{}

	// Normal
	model["title"] = testTitle
	model["value_options"] = testValueOptions
	model["is_expand"] = testTrue
	err := envOptions.ParseFromInterfaceMap(model)
	if err != nil {
		t.Fatal(err)
	}

	// title is not a string
	model = map[string]interface{}{}
	model["title"] = true
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("Title value is not a string, should be error")
	}

	// value_options is not a string slice
	model = map[string]interface{}{}
	model["value_options"] = []interface{}{true, false}
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("value_options is not a string slice, should be error")
	}

	// is_required is not a bool
	model = map[string]interface{}{}
	model["is_required"] = pointers.NewBoolPtr(testTrue)
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("is_required is not a bool, should be error")
	}

	// other_key is not supported key
	model = map[string]interface{}{}
	model["other_key"] = testTrue
	err = envOptions.ParseFromInterfaceMap(model)
	if err == nil {
		t.Fatal("other_key is not a supported key, should be error")
	}
}

func TestGetOptions(t *testing.T) {
	// Filled env
	env := EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:    pointers.NewStringPtr(testTitle),
			IsExpand: pointers.NewBoolPtr(testFalse),
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
		OptionsKey: map[interface{}]interface{}{
			"title": testTitle,
			"test":  testDescription,
		},
	}
	_, err = env.GetOptions()
	if err == nil {
		t.Fatal(err)
	}
}

func TestNormalize(t *testing.T) {
	// Filled with map[string]interface{} options
	env := EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: map[interface{}]interface{}{
			"title":         testTitle,
			"description":   testDescription,
			"summary":       testSummary,
			"value_options": testValueOptions,
			"is_required":   testTrue,
		},
	}

	err := env.Normalize()
	if err != nil {
		t.Fatal(err)
	}

	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Title == nil || *opts.Title != testTitle {
		t.Fatal("Title is nil, or not correct")
	}
	if opts.Description == nil || *opts.Description != testDescription {
		t.Fatal("Description is nil, or not correct")
	}
	if opts.Summary == nil || *opts.Summary != testSummary {
		t.Fatal("Summary is nil, or not correct")
	}
	if len(opts.ValueOptions) != len(testValueOptions) {
		t.Fatal("ValueOptions element num is not correct, or not correct")
	}
	if opts.IsRequired == nil || *opts.IsRequired != testTrue {
		t.Fatal("IsRequired is nil, or not correct")
	}

	// Filled with EnvironmentItemOptionsModel options
	env = EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:        pointers.NewStringPtr(testTitle),
			Description:  pointers.NewStringPtr(testDescription),
			Summary:      pointers.NewStringPtr(testSummary),
			ValueOptions: testValueOptions,
			IsRequired:   pointers.NewBoolPtr(testTrue),
		},
	}

	err = env.Normalize()
	if err != nil {
		t.Fatal(err)
	}

	opts, err = env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Title == nil || *opts.Title != testTitle {
		t.Fatal("Title is nil, or not correct")
	}
	if opts.Description == nil || *opts.Description != testDescription {
		t.Fatal("Description is nil, or not correct")
	}
	if opts.Summary == nil || *opts.Summary != testSummary {
		t.Fatal("Summary is nil, or not correct")
	}
	if len(opts.ValueOptions) != len(testValueOptions) {
		t.Fatal("ValueOptions element num is not correct")
	}
	if opts.IsRequired == nil || *opts.IsRequired != testTrue {
		t.Fatal("IsRequired is nil, or not correct")
	}

	// Empty options
	env = EnvironmentItemModel{
		testKey: testValue,
	}

	err = env.Normalize()
	if err != nil {
		t.Fatal(err)
	}

	opts, err = env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Title != nil {
		t.Fatal("Title is not nil")
	}
	if opts.Description != nil {
		t.Fatal("Description is not nil")
	}
	if opts.Summary != nil {
		t.Fatal("Summary is not nil")
	}
	if len(opts.ValueOptions) != 0 {
		t.Fatal("ValueOptions element num is not correct")
	}
	if opts.IsRequired != nil {
		t.Fatal("IsRequired is not nil")
	}
}

func TestFillMissingDefaults(t *testing.T) {
	// Empty env
	env := EnvironmentItemModel{
		testKey: testValue,
	}
	err := env.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}
	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	// required
	//  if opts.Title == nil || *opts.Title != "" {
	//  	t.Fatal("Failed to fill Title default value")
	//  }
	if opts.Description == nil || *opts.Description != "" {
		t.Fatal("Failed to fill Description default value")
	}
	if opts.Summary == nil || *opts.Summary != "" {
		t.Fatal("Failed to fill Summary default value")
	}
	if opts.IsRequired == nil || *opts.IsRequired != DefaultIsRequired {
		t.Fatal("Failed to fill IsRequired default value")
	}
	if opts.IsExpand == nil || *opts.IsExpand != DefaultIsExpand {
		t.Fatal("Failed to fill IsExpand default value")
	}
	if opts.IsDontChangeValue == nil || *opts.IsDontChangeValue != DefaultIsDontChangeValue {
		t.Fatal("Failed to fill IsDontChangeValue default value")
	}

	// Filled env
	env = EnvironmentItemModel{
		testKey: testValue,
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			Summary:           pointers.NewStringPtr(testSummary),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testTrue),
			IsDontChangeValue: pointers.NewBoolPtr(testFalse),
		},
	}
	err = env.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}
	opts, err = env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if opts.Title == nil || *opts.Title != testTitle {
		t.Fatal("Title is nil, or not correct")
	}
	if opts.Description == nil || *opts.Description != testDescription {
		t.Fatal("Description is nil, or not correct")
	}
	if opts.Summary == nil || *opts.Summary != testSummary {
		t.Fatal("Summary is nil, or not correct")
	}
	if len(opts.ValueOptions) != len(testValueOptions) {
		t.Fatal("ValueOptions element num is not correct")
	}
	if opts.IsRequired == nil || *opts.IsRequired != testTrue {
		t.Fatal("IsRequired is nil, or not correct")
	}
	if opts.IsExpand == nil || *opts.IsExpand != testTrue {
		t.Fatal("IsExpand is nil, or not correct")
	}
	if opts.IsDontChangeValue == nil || *opts.IsDontChangeValue != testFalse {
		t.Fatal("IsDontChangeValue is nil, or not correct")
	}
}

func TestValidate(t *testing.T) {
	// No key-value
	env := EnvironmentItemModel{
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			Summary:           pointers.NewStringPtr(testSummary),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testTrue),
			IsDontChangeValue: pointers.NewBoolPtr(testFalse),
		},
	}
	err := env.Validate()
	if err == nil {
		t.Fatal("Should be invalid env, no key-value")
	}

	// Empty key
	env = EnvironmentItemModel{
		"": testValue,
	}
	err = env.Validate()
	if err == nil {
		t.Fatal("Should be invalid env, no empty key")
	}

	// Valid env
	env = EnvironmentItemModel{
		testKey: testValue,
	}
	err = env.Validate()
	if err != nil {
		t.Fatal(err)
	}
}
