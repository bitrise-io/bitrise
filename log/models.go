package log

// StepStartedParams ...
type StepStartedParams struct {
	ExecutionId string `json:"uuid"`
	Position    int    `json:"idx"`
	IdVersion   string `json:"id_version"`
	Id          string `json:"id"`
	Version     string `json:"version"`
	Title       string `json:"title"`
	Collection  string `json:"collection"`
	Toolkit     string `json:"toolkit"`
	StartTime   string `json:"-"` // This value is only needed for the console logging.
}
