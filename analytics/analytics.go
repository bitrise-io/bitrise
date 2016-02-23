package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
)

//=======================================
// Consts
//=======================================

const analyticsURL = "https://bitrise-stats.herokuapp.com/save"

//=======================================
// Models
//=======================================

// AnonymizedUsageModel ...
type AnonymizedUsageModel struct {
	ID      string        `json:"step"`
	Version string        `json:"version"`
	RunTime time.Duration `json:"duration"`
	Error   bool          `json:"error,omitempty"`
}

// AnonymizedUsageGroupModel ...
type AnonymizedUsageGroupModel struct {
	Steps []AnonymizedUsageModel `json:"steps"`
}

//=======================================
// Main
//=======================================

// SendAnonymizedAnalytics ...
func SendAnonymizedAnalytics(buildRunResults models.BuildRunResultsModel) error {
	bitrise.PrintAnonymizedUsage()

	orderedResults := buildRunResults.OrderedResults()

	anonymizedUsageGroup := AnonymizedUsageGroupModel{}
	for _, stepRunResult := range orderedResults {
		anonymizedUsageData := AnonymizedUsageModel{
			ID:      stepRunResult.StepInfo.ID,
			Version: stepRunResult.StepInfo.Version,
			RunTime: stepRunResult.RunTime,
			Error:   stepRunResult.Status != models.StepRunStatusCodeSuccess,
		}

		anonymizedUsageGroup.Steps = append(anonymizedUsageGroup.Steps, anonymizedUsageData)
	}

	data, err := json.Marshal(anonymizedUsageGroup)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", analyticsURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("Failed to close response body, error: %#v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 210 {
		return fmt.Errorf("Sending analytics data, failed with status code: %d", resp.StatusCode)
	}

	return nil
}
