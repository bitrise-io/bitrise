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
