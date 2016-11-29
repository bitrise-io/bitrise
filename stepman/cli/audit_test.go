package cli

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/models"
)

// Test - Stepman audit step
// Checks if step Source.Commit meets the git commit hash of realese version
// 'auditStep(...)' calls 'cmdex.GitCloneTagOrBranchAndValidateCommitHash(...)', which method validates the commit hash
func TestValidateStepCommitHash(t *testing.T) {
	// Slack step - valid hash
	stepSlack := models.StepModel{
		Title:       pointers.NewStringPtr("hash_test"),
		Summary:     pointers.NewStringPtr("summary"),
		Website:     pointers.NewStringPtr("website"),
		PublishedAt: pointers.NewTimePtr(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Source: &models.StepSourceModel{
			Git:    "https://github.com/bitrise-io/steps-slack-message.git",
			Commit: "756f39f76f94d525aaea2fc2d0c5a23799f8ec97",
		},
	}
	if err := auditStepModelBeforeSharePullRequest(stepSlack, "slack", "2.1.0"); err != nil {
		t.Fatal("Step audit failed:", err)
	}

	// Slack step - invalid hash
	stepSlack.Source.Commit = "should fail commit"
	if err := auditStepModelBeforeSharePullRequest(stepSlack, "slack", "2.1.0"); err == nil {
		t.Fatal("Step audit should fail")
	}

	// Slack step - empty hash
	stepSlack.Source.Commit = ""
	if err := auditStepModelBeforeSharePullRequest(stepSlack, "slack", "2.1.0"); err == nil {
		t.Fatal("Step audit should fail")
	}
}
