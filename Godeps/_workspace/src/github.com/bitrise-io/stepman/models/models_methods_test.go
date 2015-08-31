package models

import (
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
)

const (
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

	testTitle       = "test_title"
	testDescription = "test_description"
	testTrue        = true
	testFalse       = false
)

var (
	testValueOptions = []string{testKey2, testValue2}
)

func TestValidateStep(t *testing.T) {
	step := StepModel{
		Title:   pointers.NewStringPtr("title"),
		Summary: pointers.NewStringPtr("summary"),
		Website: pointers.NewStringPtr("website"),
		Source: StepSourceModel{
			Git:    "https://github.com/bitrise-io/bitrise.git",
			Commit: "1e1482141079fc12def64d88cb7825b8f1cb1dc3",
		},
	}

	if err := step.ValidateStep(true); err != nil {
		t.Fatal(err)
	}

	step.Title = nil
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: no Title defined")
	}
	step.Title = new(string)

	*step.Title = ""
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: empty Title")
	}

	step.Description = nil
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: no Description defined")
	}
	step.Description = new(string)

	*step.Description = ""
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: empty Description")
	}

	step.Website = nil
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: no Website defined")
	}
	step.Website = new(string)

	*step.Website = ""
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: empty Website")
	}

	step.Source.Git = ""
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: empty Source.Git")
	}

	step.Source.Git = "git@github.com:bitrise-io/bitrise.git"
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: Source.Git has invalid prefix")
	}

	step.Source.Git = "https://github.com/bitrise-io/bitrise"
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: Source.Git has invalid suffix")
	}

	step.Source.Commit = ""
	if err := step.ValidateStep(true); err == nil {
		t.Fatal("Invalid step: empty Source.Commit")
	}
}

func TestValidateStepInputOutputModel(t *testing.T) {
	// Filled env
	env := envmanModels.EnvironmentItemModel{
		testKey: testValue,
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
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
			Title:             pointers.NewStringPtr(testTitle),
			Description:       pointers.NewStringPtr(testDescription),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
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
			Description:       pointers.NewStringPtr(testDescription),
			ValueOptions:      testValueOptions,
			IsRequired:        pointers.NewBoolPtr(testTrue),
			IsExpand:          pointers.NewBoolPtr(testFalse),
			IsDontChangeValue: pointers.NewBoolPtr(testTrue),
		},
	}

	err = ValidateStepInputOutputModel(env)
	if err == nil {
		t.Fatal("Empty Title, should fail")
	}
}

func TestFillMissingDefaults(t *testing.T) {
	title := "name 1"
	// desc := "desc 1"
	website := "web/1"
	git := "https://git.url"
	// fork := "fork/1"

	step := StepModel{
		Title:   pointers.NewStringPtr(title),
		Website: pointers.NewStringPtr(website),
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
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	step := StepModel{
		Title:         pointers.NewStringPtr(title),
		Description:   pointers.NewStringPtr(desc),
		Website:       pointers.NewStringPtr(website),
		SourceCodeURL: pointers.NewStringPtr(fork),
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: pointers.NewBoolPtr(defaultIsRequiresAdminUser),
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
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	// Zip & git download locations
	step := StepModel{
		Title:         pointers.NewStringPtr(title),
		Description:   pointers.NewStringPtr(desc),
		Website:       pointers.NewStringPtr(website),
		SourceCodeURL: pointers.NewStringPtr(fork),
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: pointers.NewBoolPtr(defaultIsRequiresAdminUser),
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
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	step := StepModel{
		Title:         pointers.NewStringPtr(title),
		Description:   pointers.NewStringPtr(desc),
		Website:       pointers.NewStringPtr(website),
		SourceCodeURL: pointers.NewStringPtr(fork),
		Source: StepSourceModel{
			Git: git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: pointers.NewBoolPtr(defaultIsRequiresAdminUser),
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
