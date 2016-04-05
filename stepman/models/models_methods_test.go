package models

import (
	"testing"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	step := StepModel{
		Title:       pointers.NewStringPtr("title"),
		Summary:     pointers.NewStringPtr("summary"),
		Website:     pointers.NewStringPtr("website"),
		PublishedAt: pointers.NewTimePtr(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Source: StepSourceModel{
			Git:    "https://github.com/bitrise-io/bitrise.git",
			Commit: "1e1482141079fc12def64d88cb7825b8f1cb1dc3",
		},
	}

	require.Equal(t, nil, step.Audit())

	step.Title = nil
	require.NotEqual(t, nil, step.Audit())

	step.Title = new(string)
	*step.Title = ""
	require.NotEqual(t, nil, step.Audit())

	step.PublishedAt = nil
	require.NotEqual(t, nil, step.Audit())
	step.PublishedAt = new(time.Time)

	*step.PublishedAt = time.Time{}
	require.NotEqual(t, nil, step.Audit())

	step.Description = nil
	require.NotEqual(t, nil, step.Audit())
	step.Description = new(string)

	*step.Description = ""
	require.NotEqual(t, nil, step.Audit())

	step.Website = nil
	require.NotEqual(t, nil, step.Audit())
	step.Website = new(string)

	*step.Website = ""
	require.NotEqual(t, nil, step.Audit())

	step.Source.Git = ""
	require.NotEqual(t, nil, step.Audit())

	step.Source.Git = "git@github.com:bitrise-io/bitrise.git"
	require.NotEqual(t, nil, step.Audit())

	step.Source.Git = "https://github.com/bitrise-io/bitrise"
	require.NotEqual(t, nil, step.Audit())

	step.Source.Commit = ""
	require.NotEqual(t, nil, step.Audit())
}

func TestValidateStepInputOutputModel(t *testing.T) {
	// Filled env
	env := envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step := StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.NoError(t, step.ValidateInputAndOutputEnvs(true))

	// Empty key
	env = envmanModels.EnvironmentItemModel{
		"": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step = StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.Error(t, step.ValidateInputAndOutputEnvs(true))

	// Title is empty
	env = envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Description:       pointers.NewStringPtr("test_description"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step = StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.Error(t, step.ValidateInputAndOutputEnvs(true))
}

func TestFillMissingDefaults(t *testing.T) {
	title := "name 1"
	// "desc 1" := ""desc 1" 1"
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

	require.Equal(t, nil, step.FillMissingDefaults())

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
		Title:         pointers.NewStringPtr("name 1"),
		Description:   pointers.NewStringPtr("desc 1"),
		Website:       pointers.NewStringPtr("web/1"),
		SourceCodeURL: pointers.NewStringPtr("fork/1"),
		Source: StepSourceModel{
			Git: "https://git.url",
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

	_, found := collection.GetStep("step", "1.0.0")
	require.Equal(t, true, found)
}

func TestGetDownloadLocations(t *testing.T) {
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	// Zip & git download locations
	step := StepModel{
		Title:         pointers.NewStringPtr("name 1"),
		Description:   pointers.NewStringPtr("desc 1"),
		Website:       pointers.NewStringPtr("web/1"),
		SourceCodeURL: pointers.NewStringPtr("fork/1"),
		Source: StepSourceModel{
			Git: "https://git.url",
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
	require.Equal(t, nil, err)

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
			if location.Src != "https://git.url" {
				t.Fatalf("Incorrect git location (%s)", location.Src)
			}
			gitFount = true
			gitIdx = idx
		}
	}

	require.Equal(t, true, zipFound)
	require.Equal(t, true, gitFount)
	if gitIdx < zipIdx {
		t.Fatal("Incorrect download locations order")
	}
}

func TestGetLatestStepVersion(t *testing.T) {
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	step := StepModel{
		Title:         pointers.NewStringPtr("name 1"),
		Description:   pointers.NewStringPtr("desc 1"),
		Website:       pointers.NewStringPtr("web/1"),
		SourceCodeURL: pointers.NewStringPtr("fork/1"),
		Source: StepSourceModel{
			Git: "https://git.url",
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
	require.Equal(t, nil, err)
	require.Equal(t, "2.0.0", latest)
}
