package models

import (
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

var (
	testKey    = "test_key"
	testValue  = "test_value"
	testKey1   = "test_key1"
	testValue1 = "test_value1"
	testKey2   = "test_key2"
	testValue2 = "test_value2"

	title   = "name 1"
	desc    = "desc 1"
	website = "web/1"
	git     = "https://git.url"
	fork    = "fork/1"

	testTitle        = "test_title"
	testDescription  = "test_description"
	testValueOptions = []string{"test_valu_options1", "test_valu_options2"}
	testTrue         = true
	testFalse        = false
)

// Workflow
func TestValidate(t *testing.T) {
	workflow := WorkflowModel{
		BeforeRun: []string{"befor1", "befor2", "befor3"},
		AfterRun:  []string{"after1", "after2", "after3"},
	}
	err := workflow.Validate("title")
	if err != nil {
		t.Fatal(err)
	}

	// refrence  cycle
	workflow = WorkflowModel{
		BeforeRun: []string{"befor1", "befor2", "befor3"},
		AfterRun:  []string{"after1", "after2", "befor1"},
	}
	err = workflow.Validate("title")
	if err == nil {
		t.Fatal("Reference cycle, should case of error")
	}
}

func TestMergeEnvironmentWith(t *testing.T) {
	// Different keys
	diffEnv := envmanModels.EnvironmentItemModel{
		testKey: testValue,
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             &testTitle,
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}
	env := envmanModels.EnvironmentItemModel{
		testKey1: testValue,
	}

	err := MergeEnvironmentWith(&env, diffEnv)
	if err == nil {
		t.Fatal("Different keys, should case of error")
	}

	// Normal merge
	env = envmanModels.EnvironmentItemModel{
		testKey:                 testValue,
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{},
	}

	err = MergeEnvironmentWith(&env, diffEnv)
	if err != nil {
		t.Fatal(err)
	}

	options, err := env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	diffOptions, err := diffEnv.GetOptions()
	if err != nil {
		t.Fatal(err)
	}

	if *options.Title != *diffOptions.Title {
		t.Fatal("Failed to merge Title")
	}
	if *options.Description != *diffOptions.Description {
		t.Fatal("Failed to merge Description")
	}
	if len(options.ValueOptions) != len(diffOptions.ValueOptions) {
		t.Fatal("Failed to merge ValueOptions")
	}
	if *options.IsRequired != *diffOptions.IsRequired {
		t.Fatal("Failed to merge IsRequired")
	}
	if *options.IsExpand != *diffOptions.IsExpand {
		t.Fatal("Failed to merge IsExpand")
	}
	if *options.IsDontChangeValue != *diffOptions.IsDontChangeValue {
		t.Fatal("Failed to merge IsDontChangeValue")
	}
}

func TestMergeStepWith(t *testing.T) {
	// title := "name 1"
	desc := "desc 1"
	website := "web/1"
	// git := "https://git.url"
	fork := "fork/1"

	defaultTrue := true

	stepData := stepmanModels.StepModel{
		// Title:         &title,
		Description:   &desc,
		Website:       &website,
		SourceCodeURL: &fork,
		// Source: stepmanModels.StepSourceModel{
		// 	Git: &git,
		// },
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: &defaultTrue,
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []envmanModels.EnvironmentItemModel{},
	}

	diffTitle := "name 2"
	newSuppURL := "supp"
	runIfStr := `{{getenv "CI" | eq "true"}}`
	stepDiffToMerge := stepmanModels.StepModel{
		Title:      &diffTitle,
		HostOsTags: []string{"linux"},
		Source: stepmanModels.StepSourceModel{
			Git: &git,
		},
		Dependencies: []stepmanModels.DependencyModel{
			stepmanModels.DependencyModel{
				Manager: "brew",
				Name:    "test",
			},
		},
		SupportURL: &newSuppURL,
		RunIf:      &runIfStr,
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2 CHANGED",
			},
		},
	}

	mergedStepData, err := MergeStepWith(stepData, stepDiffToMerge)
	if err != nil {
		t.Fatal("Failed to convert: ", err)
	}

	t.Logf("-> MERGED Step Data: %#v\n", mergedStepData)

	if *mergedStepData.Title != "name 2" {
		t.Fatal("mergedStepData.Title incorrectly converted:", *mergedStepData.Title)
	}
	if mergedStepData.HostOsTags[0] != "linux" {
		t.Fatal("mergedStepData.HostOsTags incorrectly converted:", mergedStepData.HostOsTags)
	}
	if *mergedStepData.RunIf != `{{getenv "CI" | eq "true"}}` {
		t.Fatal("mergedStepData.RunIf incorrectly converted:", *mergedStepData.RunIf)
	}
	if len(mergedStepData.Dependencies) != 1 {
		t.Fatal("mergedStepData.Dependencies incorrectly converted:", mergedStepData.Dependencies)

	} else {
		dep := mergedStepData.Dependencies[0]
		if dep.Manager != "brew" || dep.Name != "test" {
			t.Fatal("mergedStepData.Dependencies incorrectly converted:", mergedStepData.Dependencies)
		}
	}

	//
	input0 := mergedStepData.Inputs[0]
	key0, value0, err := input0.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get key-value:", err)
	}
	if key0 != "KEY_1" {
		t.Fatal("Inputs[0].EnvKey incorrectly converted")
	}
	if value0 != "Value 1" {
		t.Fatal("Inputs[0].Value incorrectly converted")
	}

	input1 := mergedStepData.Inputs[1]
	key1, value1, err := input1.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get key-value:", err)
	}
	if key1 != "KEY_2" {
		t.Fatal("Inputs[1].EnvKey incorrectly converted")
	}
	if value1 != "Value 2 CHANGED" {
		t.Fatal("Inputs[1].Value incorrectly converted")
	}
}

func TestGetInputByKey(t *testing.T) {
	stepData := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
	}

	_, found := getInputByKey(stepData, "KEY_1")
	if found == false {
		t.Fatal("Failed to find env (KEY_1)")
	}

	_, found = getInputByKey(stepData, "KEY_3")
	if found {
		t.Fatal("(KEY_3) found, even it doesn't exist")
	}
}

func TestGetStepIDStepDataPair(t *testing.T) {
	stepData := stepmanModels.StepModel{}

	stepListItem := StepListItemModel{
		"step1": stepData,
	}

	id, _, err := GetStepIDStepDataPair(stepListItem)
	if err != nil {
		t.Fatal(err)
	}
	if id != "step1" {
		t.Fatalf("Invalid step id (%s) found", id)
	}

	stepListItem = StepListItemModel{
		"step1": stepData,
		"step2": stepData,
	}

	id, _, err = GetStepIDStepDataPair(stepListItem)
	if err == nil {
		t.Fatal("2 key-value, should case of error")
	}
}

func TestCreateStepIDDataFromString(t *testing.T) {
	// default / long / verbose ID mode
	stepCompositeIDString := "steplib-src::step-id@0.0.1"
	t.Log("stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err := CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "step-id" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// no steplib-source
	stepCompositeIDString = "step-id@0.0.1"
	t.Log("(no steplib-source, but default provided) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "default-steplib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "default-steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "step-id" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// invalid/empty step lib source, but default provided
	stepCompositeIDString = "::step-id@0.0.1"
	t.Log("(invalid/empty steplib source, but default provided) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "default-steplib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "default-steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "step-id" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// invalid/empty step lib source + no default
	stepCompositeIDString = "::step-id@0.0.1"
	t.Log("(invalid/empty steplib source) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err == nil {
		t.Fatal("Should fail to parse the ID if it contains an empty steplib-src and no default src is provided")
	}

	// no steplib-source & no default -> fail
	stepCompositeIDString = "step-id@0.0.1"
	t.Log("(no steplib-source & no default, should fail) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err == nil {
		t.Fatal("Should fail to parse the ID if it does not contain a steplib-src and no default src is provided")
	} else {
		t.Log("Expected error (ok): ", err)
	}

	// no steplib & no version, only step-id
	stepCompositeIDString = "step-id"
	t.Log("no steplib & no version, only step-id and default lib source: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-lib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "def-lib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "step-id" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// empty test
	stepCompositeIDString = ""
	t.Log("Empty stepCompositeIDString test")
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-step-src")
	if err == nil {
		t.Fatal("Should fail to parse the ID from an empty string! (at least the step-id is required)")
	} else {
		t.Log("Expected error (ok): ", err)
	}

	// special empty test
	stepCompositeIDString = "@1.0.0"
	t.Log("Empty stepCompositeIDString test with only version")
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-step-src")
	if err == nil {
		t.Fatal("Should fail to parse the ID from an empty string! (at least the step-id is required)")
	} else {
		t.Log("Expected error (ok): ", err)
	}

	//
	// ----- Local Path
	stepCompositeIDString = "path::/some/path"
	t.Log("LOCAL - stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "path" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "/some/path" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	stepCompositeIDString = "path::~/some/path/in/home"
	t.Log("LOCAL - path should be preserved as-it-is, #1 - stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "path" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "~/some/path/in/home" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	stepCompositeIDString = "path::$HOME/some/path/in/home"
	t.Log("LOCAL - path should be preserved as-it-is, #1 - stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "path" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "$HOME/some/path/in/home" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	//
	// ----- Direct git uri
	stepCompositeIDString = "git::https://github.com/bitrise-io/steps-timestamp.git@develop"
	t.Log("DIRECT-GIT - http(s) - stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "some-def-coll")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "git" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "https://github.com/bitrise-io/steps-timestamp.git" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "develop" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	stepCompositeIDString = "git::git@github.com:bitrise-io/steps-timestamp.git@develop"
	t.Log("DIRECT-GIT - ssh - stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "git" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.IDorURI != "git@github.com:bitrise-io/steps-timestamp.git" {
		t.Fatal("stepIDData.IDorURI incorrectly converted:", stepIDData.IDorURI)
	}
	if stepIDData.Version != "develop" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}
}
