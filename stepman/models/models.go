package models

import (
	log "github.com/Sirupsen/logrus"
)

// -------------------
// --- Models

// EnvironmentItemModel ...
type EnvironmentItemModel struct {
	EnvKey            string   `json:"env_key" yaml:"env_key"`
	Value             string   `json:"value" yaml:"value"`
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// StepSourceModel ...
type StepSourceModel struct {
	Git string `json:"git" yaml:"git"`
}

// StepModel ...
type StepModel struct {
	ID                  string                 `json:"id"`
	SteplibSource       string                 `json:"steplib_source"`
	VersionTag          string                 `json:"version_tag"`
	Name                string                 `json:"name" yaml:"name"`
	Description         string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string                 `json:"website" yaml:"website"`
	ForkURL             string                 `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel        `json:"source" yaml:"source"`
	HostOsTags          []string               `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string               `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string               `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser *bool                  `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []EnvironmentItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepGroupModel ...
type StepGroupModel struct {
	ID       string      `json:"id"`
	Versions []StepModel `json:"versions"`
	Latest   StepModel   `json:"latest"`
}

// StepHash ...
type StepHash map[string]StepGroupModel

// DownloadLocationModel ...
type DownloadLocationModel struct {
	Type string `json:"type"`
	Src  string `json:"src"`
}

// StepCollectionModel ...
type StepCollectionModel struct {
	FormatVersion        string                  `json:"format_version" yaml:"format_version"`
	GeneratedAtTimeStamp int64                   `json:"generated_at_timestamp" yaml:"generated_at_timestamp"`
	Steps                StepHash                `json:"steps" yaml:"steps"`
	SteplibSource        string                  `json:"steplib_source" yaml:"steplib_source"`
	DownloadLocations    []DownloadLocationModel `json:"download_locations" yaml:"download_locations"`
}

// WorkFlowModel ...
type WorkFlowModel struct {
	FormatVersion string      `json:"format_version"`
	Environments  []string    `json:"environments"`
	Steps         []StepModel `json:"steps"`
}

// -------------------
// --- Struct methods

// GetStep ...
func (collection StepCollectionModel) GetStep(id, version string) (bool, StepModel) {
	log.Debugln("-> GetStep")
	versions := collection.Steps[id].Versions
	for _, step := range versions {
		log.Debugf(" Iterating... itm: %#v\n", step)
		if step.VersionTag == version {
			return true, step
		}
	}
	return false, StepModel{}
}

// GetDownloadLocations ...
func (collection StepCollectionModel) GetDownloadLocations(step StepModel) []DownloadLocationModel {
	locations := []DownloadLocationModel{}
	for _, downloadLocation := range collection.DownloadLocations {
		switch downloadLocation.Type {
		case "zip":
			url := downloadLocation.Src + step.ID + "/" + step.VersionTag + "/step.zip"
			location := DownloadLocationModel{
				Type: downloadLocation.Type,
				Src:  url,
			}
			locations = append(locations, location)
		case "git":
			location := DownloadLocationModel{
				Type: downloadLocation.Type,
				Src:  step.Source.Git,
			}
			locations = append(locations, location)
		default:
			log.Error("[STEPMAN] - Invalid download location")
		}
	}
	return locations
}
