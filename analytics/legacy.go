package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"net/http"
	"time"

	log "github.com/bitrise-io/go-utils/v2/advancedlog"
)

var (
	analyticsServerURL = "https://bitrise-step-analytics.herokuapp.com"
	httpClient         = http.Client{
		Timeout: time.Second * 5,
	}
)

// LogMessage sends the log message to the configured analytics server
func LogMessage(logLevel string, stepID string, tag string, data map[string]interface{}, format string, v ...interface{}) {
	// Entry represents a line in a log
	e := struct {
		LogLevel string                 `json:"log_level"`
		Message  string                 `json:"message"`
		Data     map[string]interface{} `json:"data"`
	}{
		Message:  fmt.Sprintf(format, v...),
		LogLevel: logLevel,
	}

	e.Data = make(map[string]interface{})
	for k, v := range data {
		e.Data[k] = v
	}

	if v, ok := e.Data["step_id"]; ok {
		log.Printf("internal logger: data.step_id (%s) will be overriden with (%s) ", v, stepID)
	}
	if v, ok := e.Data["tag"]; ok {
		log.Printf("internal logger: data.tag (%s) will be overriden with (%s) ", v, tag)
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
