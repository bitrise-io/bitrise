package log

// StepStartedParams ...
type StepStartedParams struct {
	ExecutionId string `json:"uuid"`
	Position    int    `json:"idx"`
	Title       string `json:"title"`
	Id          string `json:"id"`
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

// StepError ...
type StepError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StepFinishedParams ...
type StepFinishedParams struct {
	ExecutionId string `json:"uuid"`
	// The status we send to the log service is string based, but it is easier to work with an int.
	InternalStatus int         `json:"-"`
	Status         string      `json:"status"`
	StatusReason   string      `json:"status_reason,omitempty"`
	Title          string      `json:"title"`
	RunTime        int64       `json:"run_time_in_ms"`
	SupportURL     string      `json:"support_url"`
	SourceCodeURL  string      `json:"source_code_url"`
	Errors         []StepError `json:"errors,omitempty"`
	// The update and deprecation fields are pointers because an empty struct is always initialised so never omitted.
	Update      *StepUpdate      `json:"update_available,omitempty"`
	Deprecation *StepDeprecation `json:"deprecation,omitempty"`
	LastStep    bool             `json:"last_step"`
}
