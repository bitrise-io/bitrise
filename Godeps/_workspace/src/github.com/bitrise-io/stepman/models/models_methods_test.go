package models

import (
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
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
	testValueOptions = []string{testKey2, testValue2}
	testTrue         = true
	testFalse        = false
)

func TestValidateStepInputOutputModel(t *testing.T) {
	// Filled env
	env := envmanModels.EnvironmentItemModel{
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

	err := ValidateStepInputOutputModel(env)
	if err != nil {
		t.Fatal(err)
	}

	// Empty key
	env = envmanModels.EnvironmentItemModel{
		"": testValue,
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             &testTitle,
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}

	err = ValidateStepInputOutputModel(env)
	if err == nil {
		t.Fatal("Empty key, should fail")
	}

	// Title is empty
	env = envmanModels.EnvironmentItemModel{
		testKey: testValue,
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Description:       &testDescription,
			ValueOptions:      testValueOptions,
			IsRequired:        &testTrue,
			IsExpand:          &testFalse,
			IsDontChangeValue: &testTrue,
		},
	}

	err = ValidateStepInputOutputModel(env)
	if err == nil {
		t.Fatal("Empty Title, should fail")
	}
}

// func TestNormalize(t *testing.T) {
// }

func TestFillMissingDefaults(t *testing.T) {
	title := "name 1"
	// desc := "desc 1"
	website := "web/1"
	git := "https://git.url"
	// fork := "fork/1"

	step := StepModel{
		Title:   &title,
		Website: &website,
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:      []string{"osx"},
		ProjectTypeTags: []string{"ios"},
		TypeTags:        []string{"test"},
	}

	err := step.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}

	if step.Description == nil || *step.Description != "" {
		t.Fatal("Description missing")
	}
	if step.SourceCodeURL == nil || *step.SourceCodeURL != "" {
		t.Fatal("SourceCodeURL missing")
	}
	if step.SupportURL == nil || *step.SupportURL != "" {
		t.Fatal("SourceCodeURL missing")
	}
	if step.IsRequiresAdminUser == nil || *step.IsRequiresAdminUser != DefaultIsRequiresAdminUser {
		t.Fatal("IsRequiresAdminUser missing")
	}
	if step.IsAlwaysRun == nil || *step.IsAlwaysRun != DefaultIsAlwaysRun {
		t.Fatal("IsAlwaysRun missing")
	}
	if step.IsSkippable == nil || *step.IsSkippable != DefaultIsSkippable {
		t.Fatal("IsSkippable missing")
	}
	if step.RunIf == nil || *step.RunIf != "" {
		t.Fatal("RunIf missing")
	}
}

func TestGetStep(t *testing.T) {
	step := StepModel{
		Title:         &title,
		Description:   &desc,
		Website:       &website,
		SourceCodeURL: &fork,
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: &DefaultIsRequiresAdminUser,
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_3": "Value 3",
			},
		},
	}

	collection := StepCollectionModel{
		FormatVersion:        "1.0.0",
		GeneratedAtTimeStamp: 0,
		Steps: StepHash{
			"step": StepGroupModel{
				Versions: map[string]StepModel{
					"1.0.0": step,
				},
			},
		},
		SteplibSource: "source",
		DownloadLocations: []DownloadLocationModel{
			DownloadLocationModel{
				Type: "zip",
				Src:  "amazon/",
			},
			DownloadLocationModel{
				Type: "git",
				Src:  "step.git",
			},
		},
	}

	step, found := collection.GetStep("step", "1.0.0")
	if !found {
		t.Fatal("Step not found (step) (1.0.0)")
	}
}

func TestGetDownloadLocations(t *testing.T) {
	// Zip & git download locations
	step := StepModel{
		Title:         &title,
		Description:   &desc,
		Website:       &website,
		SourceCodeURL: &fork,
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: &DefaultIsRequiresAdminUser,
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_3": "Value 3",
			},
		},
	}

	collection := StepCollectionModel{
		FormatVersion:        "1.0.0",
		GeneratedAtTimeStamp: 0,
		Steps: StepHash{
			"step": StepGroupModel{
				Versions: map[string]StepModel{
					"1.0.0": step,
				},
			},
		},
		SteplibSource: "source",
		DownloadLocations: []DownloadLocationModel{
			DownloadLocationModel{
				Type: "zip",
				Src:  "amazon/",
			},
			DownloadLocationModel{
				Type: "git",
				Src:  "step.git",
			},
		},
	}

	locations, err := collection.GetDownloadLocations("step", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	zipFound := false
	gitFount := false
	zipIdx := -1
	gitIdx := -1

	for idx, location := range locations {
		if location.Type == "zip" {
			if location.Src != "amazon/step/1.0.0/step.zip" {
				t.Fatalf("Incorrect zip location (%s)", location.Src)
			}
			zipFound = true
			zipIdx = idx
		} else if location.Type == "git" {
			if location.Src != git {
				t.Fatalf("Incorrect git location (%s)", location.Src)
			}
			gitFount = true
			gitIdx = idx
		}
	}

	if zipFound == false {
		t.Fatal("No zip location found")
	}
	if gitFount == false {
		t.Fatal("No zip location found")
	}
	if gitIdx < zipIdx {
		t.Fatal("Incorrect download locations order")
	}
}

func TestGetLatestStepVersion(t *testing.T) {
	step := StepModel{
		Title:         &title,
		Description:   &desc,
		Website:       &website,
		SourceCodeURL: &fork,
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: &DefaultIsRequiresAdminUser,
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_3": "Value 3",
			},
		},
	}

	collection := StepCollectionModel{
		FormatVersion:        "1.0.0",
		GeneratedAtTimeStamp: 0,
		Steps: StepHash{
			"step": StepGroupModel{
				Versions: map[string]StepModel{
					"1.0.0": step,
					"2.0.0": step,
				},
				LatestVersionNumber: "2.0.0",
			},
		},
		SteplibSource: "source",
		DownloadLocations: []DownloadLocationModel{
			DownloadLocationModel{
				Type: "zip",
				Src:  "amazon/",
			},
			DownloadLocationModel{
				Type: "git",
				Src:  "step.git",
			},
		},
	}

	latest, err := collection.GetLatestStepVersion("step")
	if err != nil {
		t.Fatal(err)
	}
	if latest != "2.0.0" {
		t.Fatalf("Latest version (%s), should be (2.0.0)", latest)
	}
}

func TestCompareVersions(t *testing.T) {
	t.Log("Trivial compare")
	if res, err := CompareVersions("1.0.0", "1.0.1"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Reverse compare")
	if res, err := CompareVersions("1.0.2", "1.0.1"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Equal compare")
	if res, err := CompareVersions("1.0.2", "1.0.2"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Missing last num in first")
	if res, err := CompareVersions("7.0", "7.0.2"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in first - eql")
	if res, err := CompareVersions("7.0", "7.0.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in second")
	if res, err := CompareVersions("7.0.2", "7.0"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in second - eql")
	if res, err := CompareVersions("7.0.0", "7.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in first")
	if res, err := CompareVersions("7", "7.0.2"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in first - eql")
	if res, err := CompareVersions("7", "7.0.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in second")
	if res, err := CompareVersions("7.0.2", "7"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in second - eql")
	if res, err := CompareVersions("7.0.0", "7"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	// specials are not handled but should not cause any issue / panic
	t.Log("Special / non number component")
	if res, err := CompareVersions("7.x.1.2.3", "7.0.1.x"); err == nil {
		t.Fatal("Not supported compare should return an error!")
	} else {
		t.Log("[expected] Failed, res:", res, "| err:", err)
	}
}
