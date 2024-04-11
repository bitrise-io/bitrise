package log

import "github.com/bitrise-io/bitrise/models"

// StepStartedParams ...
type StepStartedParams struct {
	ExecutionID string `json:"uuid"`
	Position    int    `json:"idx"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	Version     string `json:"version"`
	Collection  string `json:"collection"`
	Toolkit     string `json:"toolkit"`
	StartTime   string `json:"start_time"`
}

// StepDeprecation ...
type StepDeprecation struct {
	RemovalDate string `json:"removal_date"`
	Note        string `json:"note"`
}

// StepUpdate ...
type StepUpdate struct {
	OriginalVersion string `json:"original_version"`
	ResolvedVersion string `json:"resolved_version"`
	LatestVersion   string `json:"latest_version"`
	ReleasesURL     string `json:"release_notes"`
}

// StepFinishedParams ...
type StepFinishedParams struct {
	ExecutionID   string             `json:"uuid"`
	Status        string             `json:"status"`
	StatusReason  string             `json:"status_reason,omitempty"`
	Title         string             `json:"title"`
	RunTime       int64              `json:"run_time_in_ms"`
	SupportURL    string             `json:"support_url"`
	SourceCodeURL string             `json:"source_code_url"`
	Errors        []models.StepError `json:"errors,omitempty"`
	// The update and deprecation fields are pointers because an empty struct is always initialised so never omitted.
	Update      *StepUpdate      `json:"update_available,omitempty"`
	Deprecation *StepDeprecation `json:"deprecation,omitempty"`
	LastStep    bool             `json:"last_step"`
}
