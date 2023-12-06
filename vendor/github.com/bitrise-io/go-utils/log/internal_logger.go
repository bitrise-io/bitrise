package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	analyticsServerURL = "https://step-analytics.bitrise.io"
	httpClient         = http.Client{
		Timeout: time.Second * 5,
	}
)

// Entry represents a line in a log
type Entry struct {
	LogLevel string                 `json:"log_level"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data"`
}

// SetAnalyticsServerURL updates the the analytics server collecting the
// logs. It is intended for use during tests. Warning: current implementation
// is not thread safe, do not call the function during runtime.
func SetAnalyticsServerURL(url string) {
	analyticsServerURL = url
}

// Internal sends the log message to the configured analytics server
func rprintf(logLevel string, stepID string, tag string, data map[string]interface{}, format string, v ...interface{}) {
	e := Entry{
		Message:  fmt.Sprintf(format, v...),
		LogLevel: logLevel,
	}

	e.Data = make(map[string]interface{})
	for k, v := range data {
		e.Data[k] = v
	}

	if v, ok := e.Data["step_id"]; ok {
		fmt.Printf("internal logger: data.step_id (%s) will be overriden with (%s) ", v, stepID)
	}
	if v, ok := e.Data["tag"]; ok {
		fmt.Printf("internal logger: data.tag (%s) will be overriden with (%s) ", v, tag)
	}

	e.Data["step_id"] = stepID
	e.Data["tag"] = tag

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(e); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequest(http.MethodPost, analyticsServerURL+"/logs", &b)
	if err != nil {
		// deliberately not writing into users log
		return
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	if _, err := httpClient.Do(req); err != nil {
		// deliberately not writing into users log
		return
	}

}
