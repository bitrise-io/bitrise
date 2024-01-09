package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggerMapItemValidate(t *testing.T) {
	t.Log("utility workflow triggered - Warning")
	{
		configStr := `
format_version: 1.3.1
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

trigger_map:
- push_branch: "/release"
  workflow: _deps-update

workflows:
  _deps-update:
`

		config, err := configModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)

		warnings, err := config.Validate()
		require.NoError(t, err)
		require.Equal(t, []string{"workflow (_deps-update) defined in trigger item (push_branch: /release -> workflow: _deps-update), but utility workflows can't be triggered directly"}, warnings)
	}

	t.Log("pipeline not exists")
	{
		configStr := `
format_version: 1.3.1
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

trigger_map:
- push_branch: "/release"
  pipeline: release

pipelines:
  primary:
    stages:
    - ci-stage: {}

stages:
  ci-stage:
    workflows:
    - ci: {}

workflows:
  ci:
`

		config, err := configModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)

		_, err = config.Validate()
		require.EqualError(t, err, "pipeline (release) defined in trigger item (push_branch: /release -> pipeline: release), but does not exist")
	}

	t.Log("workflow not exists")
	{
		configStr := `
format_version: 1.3.1
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

trigger_map:
- push_branch: "/release"
  workflow: release

workflows:
  ci:
`

		config, err := configModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)

		_, err = config.Validate()
		require.EqualError(t, err, "workflow (release) defined in trigger item (push_branch: /release -> workflow: release), but does not exist")
	}

	t.Log("it validates deprecated trigger item with triggered pipeline")
	{
		item := TriggerMapItemModel{
			Pattern:    "*",
			PipelineID: "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it validates deprecated trigger item with triggered workflow")
	{
		item := TriggerMapItemModel{
			Pattern:    "*",
			WorkflowID: "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it fails for invalid deprecated trigger item - pipeline & workflow both defined")
	{
		item := TriggerMapItemModel{
			Pattern:    "*",
			PipelineID: "pipeline-1",
			WorkflowID: "workflow-1",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for invalid deprecated trigger item - missing pipeline & workflow")
	{
		item := TriggerMapItemModel{
			Pattern: "*",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for invalid deprecated trigger item - missing pattern")
	{
		item := TriggerMapItemModel{
			Pattern:    "",
			WorkflowID: "primary",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it validates code-push trigger item with triggered pipeline")
	{
		item := TriggerMapItemModel{
			PushBranch: "*",
			PipelineID: "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it validates code-push trigger item with triggered workflow")
	{
		item := TriggerMapItemModel{
			PushBranch: "*",
			WorkflowID: "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it fails for invalid code-push trigger item - missing push-branch")
	{
		item := TriggerMapItemModel{
			PushBranch: "",
			WorkflowID: "primary",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for invalid code-push trigger item - missing pipeline & workflow")
	{
		item := TriggerMapItemModel{
			PushBranch: "*",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it validates pull-request trigger item with triggered pipeline")
	{
		item := TriggerMapItemModel{
			PullRequestSourceBranch: "feature/",
			PipelineID:              "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it validates pull-request trigger item with triggered workflow")
	{
		item := TriggerMapItemModel{
			PullRequestSourceBranch: "feature/",
			WorkflowID:              "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it validates pull-request trigger item with triggered pipeline")
	{
		item := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			PipelineID:              "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it validates pull-request trigger item with triggered workflow")
	{
		item := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		require.NoError(t, item.Validate())
	}

	t.Log("it fails for invalid pull-request trigger item - missing pipeline & workflow")
	{
		item := TriggerMapItemModel{
			PullRequestTargetBranch: "*",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for invalid pull-request trigger item - missing pipeline & workflow")
	{
		item := TriggerMapItemModel{
			PullRequestSourceBranch: "",
			PullRequestTargetBranch: "",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for mixed trigger item")
	{
		item := TriggerMapItemModel{
			PushBranch:              "master",
			PullRequestSourceBranch: "feature/*",
			PullRequestTargetBranch: "",
			WorkflowID:              "primary",
		}
		require.Error(t, item.Validate())
	}

	t.Log("it fails for mixed trigger item")
	{
		item := TriggerMapItemModel{
			PushBranch: "master",
			Pattern:    "*",
			WorkflowID: "primary",
		}
		require.Error(t, item.Validate())
	}
}

func TestMatchWithParamsCodePushItem(t *testing.T) {
	t.Log("The following patterns are all matches")
	{
		for aPattern, aPushBranch := range map[string]string{
			"feature":   "feature",
			"feature/*": "feature/login",
			"feature**": "feature",
			"*feature":  "feature",
			"**feature": "feature",
			"*":         "feature",
		} {
			pushBranch := aPushBranch
			prSourceBranch := ""
			prTargetBranch := ""
			tag := ""

			item := TriggerMapItemModel{
				PushBranch: aPattern,
				WorkflowID: "primary",
			}
			match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
			require.NoError(t, err)
			require.Equal(t, true, match, "(pattern: %s) (branch: %s)", aPattern, aPushBranch)
		}
	}

	t.Log("code-push against code-push type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PushBranch: "deploy",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			WorkflowID:              "test",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}
}

func TestMatchWithParamsPrTypeItem(t *testing.T) {
	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := "feature/login"
		prTargetBranch := "develop"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "feature/*",
			PullRequestTargetBranch: "develop",
			WorkflowID:              "test",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "*",
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := "deploy_1_0_0"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestTargetBranch: "deploy_*",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("pr against pr type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "deploy",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("pr against pr type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "feature/*",
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("pr against push type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := "master"
		tag := ""

		item := TriggerMapItemModel{
			PushBranch: "master",
			WorkflowID: "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}
}

func TestMatchWithParamsTagTypeItem(t *testing.T) {
	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		item := TriggerMapItemModel{
			Tag:        "0.9.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		item := TriggerMapItemModel{
			Tag:        "0.9.0",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			Tag:        "0.9.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			Tag:        "1.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("tag against push type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			PushBranch: "master",
			WorkflowID: "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}
}

func TestTriggerMapItemModelString(t *testing.T) {
	t.Log("triggering pipeline")
	{
		item := TriggerMapItemModel{
			PushBranch: "master",
			PipelineID: "pipeline-1",
		}
		require.Equal(t, "push_branch: master -> pipeline: pipeline-1", item.String(true))
		require.Equal(t, "push_branch: master", item.String(false))
	}

	t.Log("push event")
	{
		item := TriggerMapItemModel{
			PushBranch: "master",
			WorkflowID: "ci",
		}
		require.Equal(t, "push_branch: master -> workflow: ci", item.String(true))
		require.Equal(t, "push_branch: master", item.String(false))
	}

	t.Log("pull request event")
	{
		prSourceItem := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			WorkflowID:              "ci",
		}
		require.Equal(t, "pull_request_source_branch: develop -> workflow: ci", prSourceItem.String(true))
		require.Equal(t, "pull_request_source_branch: develop", prSourceItem.String(false))

		prTargetItem := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			WorkflowID:              "ci",
		}
		require.Equal(t, "pull_request_target_branch: master -> workflow: ci", prTargetItem.String(true))
		require.Equal(t, "pull_request_target_branch: master", prTargetItem.String(false))

		prItem := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "master",
			WorkflowID:              "ci",
		}
		require.Equal(t, "pull_request_source_branch: develop && pull_request_target_branch: master -> workflow: ci", prItem.String(true))
		require.Equal(t, "pull_request_source_branch: develop && pull_request_target_branch: master", prItem.String(false))
	}

	t.Log("tag event")
	{
		item := TriggerMapItemModel{
			Tag:        "0.9.0",
			WorkflowID: "release",
		}
		require.Equal(t, "tag: 0.9.0 -> workflow: release", item.String(true))
		require.Equal(t, "tag: 0.9.0", item.String(false))
	}

	t.Log("deprecated type")
	{
		prNotAllowedItem := TriggerMapItemModel{
			Pattern:              "master",
			IsPullRequestAllowed: false,
			WorkflowID:           "ci",
		}
		require.Equal(t, "pattern: master && is_pull_request_allowed: false -> workflow: ci", prNotAllowedItem.String(true))
		require.Equal(t, "pattern: master && is_pull_request_allowed: false", prNotAllowedItem.String(false))

		prAllowedItem := TriggerMapItemModel{
			Pattern:              "master",
			IsPullRequestAllowed: true,
			WorkflowID:           "ci",
		}
		require.Equal(t, "pattern: master && is_pull_request_allowed: true -> workflow: ci", prAllowedItem.String(true))
		require.Equal(t, "pattern: master && is_pull_request_allowed: true", prAllowedItem.String(false))
	}

	t.Log("mixed")
	{
		item := TriggerMapItemModel{
			PushBranch:              "master",
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "master",
			Tag:                     "0.9.0",
			Pattern:                 "*",
			IsPullRequestAllowed:    true,
			WorkflowID:              "ci",
		}
		require.Equal(t, "push_branch: master pull_request_source_branch: develop && pull_request_target_branch: master tag: 0.9.0 pattern: * && is_pull_request_allowed: true -> workflow: ci", item.String(true))
		require.Equal(t, "push_branch: master pull_request_source_branch: develop && pull_request_target_branch: master tag: 0.9.0 pattern: * && is_pull_request_allowed: true", item.String(false))
	}
}

func TestTriggerEventType(t *testing.T) {
	t.Log("it determins trigger event type")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypeCodePush, event)
	}

	t.Log("it determins trigger event type")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypePullRequest, event)
	}

	t.Log("it determins trigger event type")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := "master"
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypePullRequest, event)
	}

	t.Log("it determins trigger event type")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypeTag, event)
	}

	t.Log("it fails without inputs")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := "master"
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}
}
