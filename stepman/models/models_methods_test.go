package models

import (
	"reflect"
	"testing"
	"time"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	step := StepModel{
		Title:       pointers.NewStringPtr("title"),
		Summary:     pointers.NewStringPtr("summary"),
		Website:     pointers.NewStringPtr("website"),
		PublishedAt: pointers.NewTimePtr(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Source: &StepSourceModel{
			Git:    "https://github.com/bitrise-io/bitrise.git",
			Commit: "1e1482141079fc12def64d88cb7825b8f1cb1dc3",
		},
	}

	require.Equal(t, nil, step.Audit())

	step.Title = nil
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'title' property")

	step.Title = new(string)
	*step.Title = ""
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'title' property")
	*step.Title = "title"

	step.PublishedAt = nil
	require.NotEqual(t, nil, step.Audit())
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'PublishedAt' property")
	step.PublishedAt = new(time.Time)

	*step.PublishedAt = time.Time{}
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'PublishedAt' property")
	step.PublishedAt = pointers.NewTimePtr(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC))

	step.Website = nil
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'website' property")

	step.Website = new(string)
	*step.Website = ""
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'website' property")
	*step.Website = "website"

	step.Source.Git = ""
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'source.git' property")
	step.Source.Git = "git"

	step.Source.Git = "git@github.com:bitrise-io/bitrise.git"
	require.EqualError(t, step.Audit(), "invalid step: step source should start with http:// or https://")

	step.Source.Git = "https://github.com/bitrise-io/bitrise"
	require.EqualError(t, step.Audit(), "invalid step: step source should end with .git")
	step.Source.Git = "https://github.com/bitrise-io/bitrise.git"

	step.Source.Commit = ""
	require.EqualError(t, step.Audit(), "invalid step: missing or empty required 'source.commit' property")
	step.Source.Commit = "1e1482141079fc12def64d88cb7825b8f1cb1dc3"

	step.NoOutputTimeout = new(int)
	*step.NoOutputTimeout = -1
	require.EqualError(t, step.Audit(), "invalid step: 'no_output_timeout' is less then 0")

	step.Timeout = new(int)
	*step.Timeout = -1
	require.EqualError(t, step.Audit(), "invalid step: timeout less then 0")
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

	// IsSensitive is true but IsExpand is not
	env = envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsSensitive:       pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step = StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.Error(t, step.ValidateInputAndOutputEnvs(true))

	// IsSensitive is not set
	env = envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
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

	require.NoError(t, step.ValidateInputAndOutputEnvs(true))

	// IsSensitive is set to false
	env = envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsSensitive:       pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step = StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.NoError(t, step.ValidateInputAndOutputEnvs(true))

	// IsExpand not set and IsSensitive set
	env = envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsSensitive:       pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(true),
		},
	}

	step = StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{env},
	}

	require.NoError(t, step.ValidateInputAndOutputEnvs(true))
}

func TestFillMissingDefaults(t *testing.T) {
	title := "name 1"
	website := "web/1"
	git := "https://git.url"

	step := StepModel{
		Title:   pointers.NewStringPtr(title),
		Website: pointers.NewStringPtr(website),
		Source: &StepSourceModel{
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
	if step.Timeout == nil || *step.Timeout != 0 {
		t.Fatal("Timeout missing")
	}
	if step.NoOutputTimeout != nil {
		t.Fatalf("No output timeout has a default value")
	}
}

func TestGetStep(t *testing.T) {
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	step := StepModel{
		Title:         pointers.NewStringPtr("name 1"),
		Description:   pointers.NewStringPtr("desc 1"),
		Website:       pointers.NewStringPtr("web/1"),
		SourceCodeURL: pointers.NewStringPtr("fork/1"),
		Source: &StepSourceModel{
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

	_, stepFound, versionFound := collection.GetStep("step", "1.0.0")
	require.Equal(t, true, (stepFound && versionFound))
}

func TestGetDownloadLocations(t *testing.T) {
	defaultIsRequiresAdminUser := DefaultIsRequiresAdminUser

	// Zip & git download locations
	step := StepModel{
		Title:         pointers.NewStringPtr("name 1"),
		Description:   pointers.NewStringPtr("desc 1"),
		Website:       pointers.NewStringPtr("web/1"),
		SourceCodeURL: pointers.NewStringPtr("fork/1"),
		Source: &StepSourceModel{
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
		switch location.Type {
		case "zip":
			if location.Src != "amazon/step/1.0.0/step.zip" {
				t.Fatalf("Incorrect zip location (%s)", location.Src)
			}
			zipFound = true
			zipIdx = idx
		case "git":
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
		Source: &StepSourceModel{
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

func Test_BrewDepModel_GetBinaryName(t *testing.T) {
	require.Equal(t, "", BrewDepModel{}.GetBinaryName())
	require.Equal(t, "awscli", BrewDepModel{Name: "awscli"}.GetBinaryName())
	require.Equal(t, "aws", BrewDepModel{Name: "awscli", BinName: "aws"}.GetBinaryName())
}

func Test_AptGetDepModel_GetBinaryName(t *testing.T) {
	require.Equal(t, "", AptGetDepModel{}.GetBinaryName())
	require.Equal(t, "awscli", AptGetDepModel{Name: "awscli"}.GetBinaryName())
	require.Equal(t, "aws", AptGetDepModel{Name: "awscli", BinName: "aws"}.GetBinaryName())
}

func TestStepCollectionModel_GetStepVersion(t *testing.T) {
	step := StepModel{
		Title: pointers.NewStringPtr("name 1"),
	}

	collection := StepCollectionModel{
		Steps: StepHash{
			"step": StepGroupModel{
				Versions: map[string]StepModel{
					"1.0.0": step,
					"1.1.0": step,
					"1.1.1": step,
					"1.2.0": step,
					"2.0.0": step,
				},
				LatestVersionNumber: "2.0.0",
			},
		},
	}

	type args struct {
		id      string
		version string
	}
	tests := []struct {
		name       string
		collection StepCollectionModel
		args       args
		want       StepVersionModel
		want1      bool
		want2      bool
	}{
		{
			name:       "Fix version",
			collection: collection,
			args: args{
				id:      "step",
				version: "1.0.0",
			},
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.0.0",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
			want2: true,
		},
		{
			name:       "Lock Minor version",
			collection: collection,
			args: args{
				id:      "step",
				version: "1.1",
			},
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.1.1",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
			want2: true,
		},
		{
			name:       "Lock Major ersion",
			collection: collection,
			args: args{
				id:      "step",
				version: "1",
			},
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.2.0",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := tt.collection
			got, got1, got2 := collection.GetStepVersion(tt.args.id, tt.args.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCollectionModel.GetStepVersion() got = %+v, want %+v, \n Diff: %s", got, tt.want, cmp.Diff(tt.want, got))
			}
			if got1 != tt.want1 {
				t.Errorf("StepCollectionModel.GetStepVersion() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("StepCollectionModel.GetStepVersion() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
