package analytics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
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
	if configs.OptOutUsageData == true {
		return nil
	}

	bitrise.PrintAnonymizedUsage(buildRunResults)

	orderedResults := buildRunResults.OrderedResults()

	anonymizedUsageGroup := AnonymizedUsageGroupModel{}
	for _, stepRunResult := range orderedResults {
		anonymizedUsageData := AnonymizedUsageModel{
			ID:      stepRunResult.StepInfo.ID,
			Version: stepRunResult.StepInfo.Version,
			RunTime: stepRunResult.RunTime,
			Error:   stepRunResult.Status != 0,
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
			log.Warnf("Failed to close response body, errror: %#v", err)
		}
	}()

	return nil
}
