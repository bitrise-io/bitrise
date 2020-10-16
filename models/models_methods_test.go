package models

import (
	"strings"
	"testing"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestCheckDuplicatedTriggerMapItems(t *testing.T) {
	t.Log("duplicated push - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (push_branch: master)")
	}

	t.Log("duplicated pull request - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop)")

		err = checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_target_branch: master)")

		err = checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop && pull_request_target_branch: master)")
	}

	t.Log("duplicated tag - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (tag: 0.9.0)")
	}

	t.Log("complex trigger map - no error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
		})

		require.NoError(t, err)
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

// ----------------------------
// --- Validate

// Config
func TestValidateConfig(t *testing.T) {
	t.Log("Valid bitriseData")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"pipeline1": PipelineModel{
					Stages: []StageListItemModel{
						StageListItemModel{"stage1": StageModel{}},
						StageListItemModel{"stage2": StageModel{}},
					},
				},
				"pipeline2": PipelineModel{
					Stages: []StageListItemModel{
						StageListItemModel{"stage3": StageModel{}},
						StageListItemModel{"stage4": StageModel{}},
					},
				},
			},
			Stages: map[string]StageModel{
				"stage1": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow1": WorkflowModel{}},
						WorkflowListItemModel{"workflow2": WorkflowModel{}},
					},
				},
				"stage2": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow3": WorkflowModel{}},
						WorkflowListItemModel{"workflow4": WorkflowModel{}},
					},
				},
				"stage3": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow5": WorkflowModel{}},
						WorkflowListItemModel{"workflow6": WorkflowModel{}},
					},
				},
				"stage4": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow7": WorkflowModel{}},
						WorkflowListItemModel{"workflow8": WorkflowModel{}},
					},
				},
			},
			Workflows: map[string]WorkflowModel{
				"workflow1": WorkflowModel{},
				"workflow2": WorkflowModel{},
				"workflow3": WorkflowModel{},
				"workflow4": WorkflowModel{},
				"workflow5": WorkflowModel{},
				"workflow6": WorkflowModel{},
				"workflow7": WorkflowModel{},
				"workflow8": WorkflowModel{},
			},
		}

		warnings, err := bitriseData.Validate()
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - pipeline ID empty")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"": PipelineModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "invalid pipeline ID (): empty")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - pipeline ID contains: `/`")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"pi/id": PipelineModel{
					Stages: []StageListItemModel{
						StageListItemModel{"stage1": StageModel{}},
					},
				},
			},
			Stages: map[string]StageModel{
				"stage1": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow1": WorkflowModel{}},
					},
				},
			},
			Workflows: map[string]WorkflowModel{
				"workflow1": WorkflowModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.NoError(t, err)
		require.Equal(t, 1, len(warnings))
		require.Equal(t, "invalid pipeline ID (pi/id): doesn't conform to: [A-Za-z0-9-_.]", warnings[0])
	}

	t.Log("Invalid bitriseData - pipeline does not have any stages")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"pipeline1": PipelineModel{
					Stages: []StageListItemModel{},
				},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "pipeline (pipeline1) should have at least 1 stage")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - pipeline does not have stages key defined")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"pipeline1": PipelineModel{},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "pipeline (pipeline1) should have at least 1 stage")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - pipeline stage does not exist")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Pipelines: map[string]PipelineModel{
				"pipeline1": PipelineModel{
					Stages: []StageListItemModel{
						StageListItemModel{"stage2": StageModel{}},
					},
				},
			},
			Stages: map[string]StageModel{
				"stage1": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow1": WorkflowModel{}},
					},
				},
			},
			Workflows: map[string]WorkflowModel{
				"workflow1": WorkflowModel{},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "stage (stage2) defined in pipeline (pipeline1), but does not exist")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - stage ID empty")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Stages: map[string]StageModel{
				"": StageModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "invalid stage ID (): empty")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - stage ID contains: `/`")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Stages: map[string]StageModel{
				"st/id": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow1": WorkflowModel{}},
					},
				},
			},
			Workflows: map[string]WorkflowModel{
				"workflow1": WorkflowModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.NoError(t, err)
		require.Equal(t, 1, len(warnings))
		require.Equal(t, "invalid stage ID (st/id): doesn't conform to: [A-Za-z0-9-_.]", warnings[0])
	}

	t.Log("Invalid bitriseData - stage does not have any workflows")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Stages: map[string]StageModel{
				"stage1": StageModel{
					Workflows: []WorkflowListItemModel{},
				},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "stage (stage1) should have at least 1 workflow")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - stage does not have workflows key defined")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Stages: map[string]StageModel{
				"stage1": StageModel{},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "stage (stage1) should have at least 1 workflow")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - stage workflow does not exist")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Stages: map[string]StageModel{
				"stage1": StageModel{
					Workflows: []WorkflowListItemModel{
						WorkflowListItemModel{"workflow2": WorkflowModel{}},
					},
				},
			},
			Workflows: map[string]WorkflowModel{
				"workflow1": WorkflowModel{},
			},
		}

		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "workflow (workflow2) defined in stage (stage1), but does not exist")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - workflow ID empty")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Workflows: map[string]WorkflowModel{
				"": WorkflowModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.EqualError(t, err, "invalid workflow ID (): empty")
		require.Equal(t, 0, len(warnings))
	}

	t.Log("Invalid bitriseData - workflow ID contains: `/`")
	{
		bitriseData := BitriseDataModel{
			FormatVersion: "1.4.0",
			Workflows: map[string]WorkflowModel{
				"wf/id": WorkflowModel{},
			},
		}
		warnings, err := bitriseData.Validate()
		require.NoError(t, err)
		require.Equal(t, 1, len(warnings))
		require.Equal(t, "invalid workflow ID (wf/id): doesn't conform to: [A-Za-z0-9-_.]", warnings[0])
	}
}

// Workflow
func TestValidateWorkflow(t *testing.T) {
	t.Log("before-after test")
	{
		workflow := WorkflowModel{
			BeforeRun: []string{"befor1", "befor2", "befor3"},
			AfterRun:  []string{"after1", "after2", "after3"},
		}

		warnings, err := workflow.Validate()
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))
	}

	t.Log("invalid workflow - Invalid env: more than 2 fields")
	{
		configStr := `
format_version: 1.4.0

default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    envs:
    - ENV_KEY: env_value
      opts:
        title: test_env
    title: Output Test
    steps:
    - script:
        title: Should fail
        inputs:
        - content: echo "Hello"
          BAD_KEY: value
`

		config := BitriseDataModel{}
		require.NoError(t, yaml.Unmarshal([]byte(configStr), &config))
		require.NoError(t, config.Normalize())

		warnings, err := config.Validate()
		require.Error(t, err)
		require.Equal(t, true, strings.Contains(err.Error(), "more than 2 keys specified: [BAD_KEY content opts]"))
		require.Equal(t, 0, len(warnings))
	}

	t.Log("valid workflow - Warning: duplicated inputs")
	{
		configStr := `format_version: 1.4.0

default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    steps:
    - script:
        title: Should fail
        inputs:
        - content: echo "Hello"
        - content: echo "Hello"
`

		config := BitriseDataModel{}
		require.NoError(t, yaml.Unmarshal([]byte(configStr), &config))
		require.NoError(t, config.Normalize())

		warnings, err := config.Validate()
		require.NoError(t, err)
		require.Equal(t, 1, len(warnings))
	}
}

// ----------------------------
// --- Merge

func TestMergeEnvironmentWith(t *testing.T) {
	diffEnv := envmanModels.EnvironmentItemModel{
		"test_key": "test_value",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_valu_options1", "test_valu_options2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
			IsTemplate:        pointers.NewBoolPtr(true),
		},
	}

	t.Log("Different keys")
	{
		env := envmanModels.EnvironmentItemModel{
			"test_key1": "test_value",
		}
		require.Error(t, MergeEnvironmentWith(&env, diffEnv))
	}

	t.Log("Normal merge")
	{
		env := envmanModels.EnvironmentItemModel{
			"test_key": "test_value",
			envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
				SkipIfEmpty: pointers.NewBoolPtr(true),
				Category:    pointers.NewStringPtr("test"),
			},
		}
		require.NoError(t, MergeEnvironmentWith(&env, diffEnv))

		options, err := env.GetOptions()
		require.NoError(t, err)

		diffOptions, err := diffEnv.GetOptions()
		require.NoError(t, err)

		require.Equal(t, *diffOptions.Title, *options.Title)
		require.Equal(t, *diffOptions.Description, *options.Description)
		require.Equal(t, *diffOptions.Summary, *options.Summary)
		require.Equal(t, len(diffOptions.ValueOptions), len(options.ValueOptions))
		require.Equal(t, *diffOptions.IsRequired, *options.IsRequired)
		require.Equal(t, *diffOptions.IsExpand, *options.IsExpand)
		require.Equal(t, *diffOptions.IsDontChangeValue, *options.IsDontChangeValue)
		require.Equal(t, *diffOptions.IsTemplate, *options.IsTemplate)

		require.Equal(t, true, *options.SkipIfEmpty)
		require.Equal(t, "test", *options.Category)
	}
}

func TestMergeStepWith(t *testing.T) {
	desc := "desc 1"
	summ := "sum 1"
	website := "web/1"
	fork := "fork/1"
	published := time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)

	stepData := stepmanModels.StepModel{
		Description:         pointers.NewStringPtr(desc),
		Summary:             pointers.NewStringPtr(summ),
		Website:             pointers.NewStringPtr(website),
		SourceCodeURL:       pointers.NewStringPtr(fork),
		PublishedAt:         pointers.NewTimePtr(published),
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: pointers.NewBoolPtr(true),
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []envmanModels.EnvironmentItemModel{},
	}

	diffTitle := "name 2"
	newSuppURL := "supp"
	runIfStr := ""
	stepDiffToMerge := stepmanModels.StepModel{
		Title:      pointers.NewStringPtr(diffTitle),
		HostOsTags: []string{"linux"},
		Source: &stepmanModels.StepSourceModel{
			Git: "https://git.url",
		},
		Dependencies: []stepmanModels.DependencyModel{
			stepmanModels.DependencyModel{
				Manager: "brew",
				Name:    "test",
			},
		},
		SupportURL: pointers.NewStringPtr(newSuppURL),
		RunIf:      pointers.NewStringPtr(runIfStr),
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2 CHANGED",
			},
		},
		Timeout: pointers.NewIntPtr(1),
		Toolkit: &stepmanModels.StepToolkitModel{
			Go: &stepmanModels.GoStepToolkitModel{
				PackageName: "test",
			},
		},
	}

	mergedStepData, err := MergeStepWith(stepData, stepDiffToMerge)
	require.NoError(t, err)

	require.Equal(t, "name 2", *mergedStepData.Title)
	require.Equal(t, "desc 1", *mergedStepData.Description)
	require.Equal(t, "sum 1", *mergedStepData.Summary)
	require.Equal(t, "web/1", *mergedStepData.Website)
	require.Equal(t, "fork/1", *mergedStepData.SourceCodeURL)
	require.Equal(t, true, (*mergedStepData.PublishedAt).Equal(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)))
	require.Equal(t, "linux", mergedStepData.HostOsTags[0])
	require.Equal(t, "", *mergedStepData.RunIf)
	require.Equal(t, 1, len(mergedStepData.Dependencies))
	require.Equal(t, "test", mergedStepData.Toolkit.Go.PackageName)
	require.Equal(t, 1, *mergedStepData.Timeout)

	dep := mergedStepData.Dependencies[0]
	require.Equal(t, "brew", dep.Manager)
	require.Equal(t, "test", dep.Name)

	// inputs
	input0 := mergedStepData.Inputs[0]
	key0, value0, err := input0.GetKeyValuePair()

	require.NoError(t, err)
	require.Equal(t, "KEY_1", key0)
	require.Equal(t, "Value 1", value0)

	input1 := mergedStepData.Inputs[1]
	key1, value1, err := input1.GetKeyValuePair()

	require.NoError(t, err)
	require.Equal(t, "KEY_2", key1)
	require.Equal(t, "Value 2 CHANGED", value1)
}

func TestGetInputByKey(t *testing.T) {
	stepData := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			envmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
	}

	_, found := getInputByKey(stepData, "KEY_1")
	require.Equal(t, true, found)

	_, found = getInputByKey(stepData, "KEY_3")
	require.Equal(t, false, found)
}

// ----------------------------
// --- WorkflowIDData

func TestGetWorkflowIDWorkflowDataPair(t *testing.T) {
	workflowData := WorkflowModel{}

	t.Log("valid workflowlist item")
	{
		workflowListItem := WorkflowListItemModel{
			"workflow1": workflowData,
		}

		id, _, err := GetWorkflowIDWorkflowDataPair(workflowListItem)
		require.NoError(t, err)
		require.Equal(t, "workflow1", id)
	}

	t.Log("invalid workflowlist item - more than 1 workflow")
	{
		workflowListItem := WorkflowListItemModel{
			"workflow1": workflowData,
			"workflow2": workflowData,
		}

		id, _, err := GetWorkflowIDWorkflowDataPair(workflowListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}

	t.Log("invalid workflowlist item - no workflow")
	{
		workflowListItem := WorkflowListItemModel{}

		id, _, err := GetWorkflowIDWorkflowDataPair(workflowListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}
}

// ----------------------------
// --- StageIDData

func TestGetStageIDStageDataPair(t *testing.T) {
	stageData := StageModel{}

	t.Log("valid stagelist item")
	{
		stageListItem := StageListItemModel{
			"stage1": stageData,
		}

		id, _, err := GetStageIDStageDataPair(stageListItem)
		require.NoError(t, err)
		require.Equal(t, "stage1", id)
	}

	t.Log("invalid stagelist item - more than 1 stage")
	{
		stageListItem := StageListItemModel{
			"stage1": stageData,
			"stage2": stageData,
		}

		id, _, err := GetStageIDStageDataPair(stageListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}

	t.Log("invalid stagelist item - no stage")
	{
		stageListItem := StageListItemModel{}

		id, _, err := GetStageIDStageDataPair(stageListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}
}

// ----------------------------
// --- StepIDData

func Test_StepIDData_IsUniqueResourceID(t *testing.T) {
	stepIDDataWithIDAndVersionSpecified := StepIDData{IDorURI: "stepid", Version: "version"}
	stepIDDataWithOnlyVersionSpecified := StepIDData{Version: "version"}
	stepIDDataWithOnlyIDSpecified := StepIDData{IDorURI: "stepid"}
	stepIDDataEmpty := StepIDData{}

	// Not Unique
	for _, aSourceID := range []string{"path", "git", "_", ""} {
		stepIDDataWithIDAndVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithIDAndVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyIDSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyIDSpecified.IsUniqueResourceID())

		stepIDDataEmpty.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataEmpty.IsUniqueResourceID())
	}

	for _, aSourceID := range []string{"a", "any-other-step-source", "https://github.com/bitrise-io/bitrise-steplib.git"} {
		// Only if StepLib, AND both ID and Version are defined, only then
		// this is a Unique Resource ID!
		stepIDDataWithIDAndVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, true, stepIDDataWithIDAndVersionSpecified.IsUniqueResourceID())

		// In any other case, it's not,
		// even if it's from a StepLib
		// but missing ID or version!
		stepIDDataWithOnlyVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyIDSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyIDSpecified.IsUniqueResourceID())

		stepIDDataEmpty.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataEmpty.IsUniqueResourceID())
	}
}

func TestGetStepIDStepDataPair(t *testing.T) {
	stepData := stepmanModels.StepModel{}

	t.Log("valid steplist item")
	{
		stepListItem := StepListItemModel{
			"step1": stepData,
		}

		id, _, err := GetStepIDStepDataPair(stepListItem)
		require.NoError(t, err)
		require.Equal(t, "step1", id)
	}

	t.Log("invalid steplist item - more than 1 step")
	{
		stepListItem := StepListItemModel{
			"step1": stepData,
			"step2": stepData,
		}

		id, _, err := GetStepIDStepDataPair(stepListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}

	t.Log("invalid steplist item - no step")
	{
		stepListItem := StepListItemModel{}

		id, _, err := GetStepIDStepDataPair(stepListItem)
		require.Error(t, err)
		require.Equal(t, "", id)
	}
}

func TestCreateStepIDDataFromString(t *testing.T) {
	type data struct {
		composite            string
		defaultSteplibSource string
		wantStepSrc          string
		wantStepID           string
		wantVersion          string
		wantErr              bool
		name                 string
	}

	stepLib := []data{
		{
			name:      "no steplib-source",
			composite: "step-id@0.0.1", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "invalid/empty step lib source, but default provided",
			composite: "::step-id@0.0.1", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "invalid/empty step lib source, no default",
			composite: "::step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "no steplib-source & no default, fail",
			composite: "step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "no steplib, no version, only step-id",
			composite: "step-id", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "default, long, verbose ID mode",
			composite: "steplib-src::step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "empty test",
			composite: "", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "special empty test",
			composite: "@1.0.0", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "old step",
			composite: "_::https://github.com/bitrise-io/steps-timestamp.git@1.0.0", defaultSteplibSource: "",
			wantStepSrc: "_", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "1.0.0",
			wantErr: false,
		},
	}

	path := []data{
		{
			name:      "local path",
			composite: "path::/some/path", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "/some/path", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "local path, tilde",
			composite: "path::~/some/path/in/home", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "~/some/path/in/home", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "local path, env",
			composite: "path::$HOME/some/path/in/home", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "$HOME/some/path/in/home", wantVersion: "",
			wantErr: false,
		},
	}

	git := []data{
		{
			name:      "direct git uri, https",
			composite: "git::https://github.com/bitrise-io/steps-timestamp.git@develop", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "develop",
			wantErr: false,
		},
		{
			name:      "direct git uri, ssh",
			composite: "git::git@github.com:bitrise-io/steps-timestamp.git@develop", defaultSteplibSource: "",
			wantStepSrc: "git", wantStepID: "git@github.com:bitrise-io/steps-timestamp.git", wantVersion: "develop",
			wantErr: false,
		},
		{
			name:      "direct git uri, https, no branch",
			composite: "git::https://github.com/bitrise-io/steps-timestamp.git", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "master",
			wantErr: false,
		},
		{
			name:      "direct git uri, ssh, no branch",
			composite: "git::git@github.com:bitrise-io/steps-timestamp.git", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "git@github.com:bitrise-io/steps-timestamp.git", wantVersion: "master",
			wantErr: false,
		},
	}

	for _, group := range [][]data{
		stepLib,
		git,
		path,
	} {
		for _, tt := range group {
			stepIDData, err := CreateStepIDDataFromString(tt.composite, tt.defaultSteplibSource)

			if tt.wantErr && (err == nil) {
				t.Fatal("tt.wantErr && (err == nil):", err)
			}

			require.Equal(t, tt.wantStepSrc, stepIDData.SteplibSource, tt.name)
			require.Equal(t, tt.wantStepID, stepIDData.IDorURI, tt.name)
			require.Equal(t, tt.wantVersion, stepIDData.Version, tt.name)
		}
	}
}

// ----------------------------
// --- RemoveRedundantFields

func TestRemoveEnvironmentRedundantFields(t *testing.T) {
	t.Log("Trivial remove - all fields should be default value")
	{
		env := envmanModels.EnvironmentItemModel{
			"TEST_KEY": "test_value",
			envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
				IsExpand:    pointers.NewBoolPtr(envmanModels.DefaultIsExpand),
				SkipIfEmpty: pointers.NewBoolPtr(envmanModels.DefaultSkipIfEmpty),

				Title:             pointers.NewStringPtr(""),
				Description:       pointers.NewStringPtr(""),
				Summary:           pointers.NewStringPtr(""),
				Category:          pointers.NewStringPtr(""),
				ValueOptions:      []string{},
				IsRequired:        pointers.NewBoolPtr(envmanModels.DefaultIsRequired),
				IsDontChangeValue: pointers.NewBoolPtr(envmanModels.DefaultIsDontChangeValue),
				IsTemplate:        pointers.NewBoolPtr(envmanModels.DefaultIsTemplate),

				Meta: map[string]interface{}{},
			},
		}
		require.NoError(t, removeEnvironmentRedundantFields(&env))

		options, err := env.GetOptions()
		require.NoError(t, err)

		require.Equal(t, (*bool)(nil), options.IsExpand)
		require.Equal(t, (*bool)(nil), options.SkipIfEmpty)

		require.Equal(t, (*string)(nil), options.Title)
		require.Equal(t, (*string)(nil), options.Description)
		require.Equal(t, (*string)(nil), options.Summary)
		require.Equal(t, (*string)(nil), options.Category)
		require.Equal(t, 0, len(options.ValueOptions))
		require.Equal(t, (*bool)(nil), options.IsRequired)
		require.Equal(t, (*bool)(nil), options.IsDontChangeValue)
		require.Equal(t, (*bool)(nil), options.IsTemplate)

		require.Equal(t, 0, len(options.Meta))
	}

	t.Log("Trivial don't remove - no fields should be default value")
	{
		env := envmanModels.EnvironmentItemModel{
			"TEST_KEY": "test_value",
			envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
				IsExpand:    pointers.NewBoolPtr(false),
				SkipIfEmpty: pointers.NewBoolPtr(true),

				Title:             pointers.NewStringPtr("t"),
				Description:       pointers.NewStringPtr("d"),
				Summary:           pointers.NewStringPtr("s"),
				Category:          pointers.NewStringPtr("c"),
				ValueOptions:      []string{"i"},
				IsRequired:        pointers.NewBoolPtr(true),
				IsDontChangeValue: pointers.NewBoolPtr(true),
				IsTemplate:        pointers.NewBoolPtr(true),

				Meta: map[string]interface{}{"is_expose": true},
			},
		}
		require.NoError(t, removeEnvironmentRedundantFields(&env))

		options, err := env.GetOptions()
		require.NoError(t, err)

		require.Equal(t, false, *options.IsExpand)
		require.Equal(t, true, *options.SkipIfEmpty)

		require.Equal(t, "t", *options.Title)
		require.Equal(t, "d", *options.Description)
		require.Equal(t, "s", *options.Summary)
		require.Equal(t, "c", *options.Category)
		require.Equal(t, "i", options.ValueOptions[0])
		require.Equal(t, true, *options.IsRequired)
		require.Equal(t, true, *options.IsDontChangeValue)
		require.Equal(t, true, *options.IsTemplate)

		require.Equal(t, map[string]interface{}{"is_expose": true}, options.Meta)
	}

	t.Log("No options - opts field shouldn't exist")
	{
		env := envmanModels.EnvironmentItemModel{
			"TEST_KEY": "test_value",
		}
		require.NoError(t, removeEnvironmentRedundantFields(&env))

		_, ok := env[envmanModels.OptionsKey]
		require.Equal(t, false, ok)
	}

	t.Log("Only meta options - opts field should remain")
	{
		env := envmanModels.EnvironmentItemModel{
			"TEST_KEY": "test_value",
			envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
				Meta: map[string]interface{}{
					"is_expose": true,
				},
			},
		}
		require.NoError(t, removeEnvironmentRedundantFields(&env))

		options, err := env.GetOptions()
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}{"is_expose": true}, options.Meta)
	}
}

func configModelFromYAMLBytes(configBytes []byte) (bitriseData BitriseDataModel, err error) {
	if err = yaml.Unmarshal(configBytes, &bitriseData); err != nil {
		return
	}
	return
}

func TestRemoveWorkflowRedundantFields(t *testing.T) {
	configStr := `format_version: 2
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"
project_type: ios

app:
  summary: "sum"
  envs:
  - ENV_KEY: env_value
    opts:
      is_required: true

workflows:
  target:
    envs:
    - ENV_KEY: env_value
      opts:
        title: test_env
    title: Output Test
    steps:
    - script:
        description: test
`

	config, err := configModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)

	err = config.RemoveRedundantFields()
	require.NoError(t, err)

	require.Equal(t, "2", config.FormatVersion)
	require.Equal(t, "https://github.com/bitrise-io/bitrise-steplib.git", config.DefaultStepLibSource)
	require.Equal(t, "ios", config.ProjectType)

	require.Equal(t, "", config.App.Title)
	require.Equal(t, "", config.App.Description)
	require.Equal(t, "sum", config.App.Summary)

	for _, env := range config.App.Environments {
		options, err := env.GetOptions()
		require.NoError(t, err)

		require.Nil(t, options.Title)
		require.Nil(t, options.Description)
		require.Nil(t, options.Summary)
		require.Equal(t, 0, len(options.ValueOptions))
		require.Equal(t, true, *options.IsRequired)
		require.Nil(t, options.IsExpand)
		require.Nil(t, options.IsDontChangeValue)
	}

	for _, workflow := range config.Workflows {
		require.Equal(t, "Output Test", workflow.Title)
		require.Equal(t, "", workflow.Description)
		require.Equal(t, "", workflow.Summary)

		for _, env := range workflow.Environments {
			options, err := env.GetOptions()
			require.NoError(t, err)

			require.Equal(t, "test_env", *options.Title)
			require.Nil(t, options.Description)
			require.Nil(t, options.Summary)
			require.Equal(t, 0, len(options.ValueOptions))
			require.Nil(t, options.IsRequired)
			require.Nil(t, options.IsExpand)
			require.Nil(t, options.IsDontChangeValue)
		}

		for _, stepListItem := range workflow.Steps {
			_, step, err := GetStepIDStepDataPair(stepListItem)
			require.NoError(t, err)

			require.Nil(t, step.Title)
			require.Equal(t, "test", *step.Description)
			require.Nil(t, step.Summary)
			require.Nil(t, step.Website)
			require.Nil(t, step.SourceCodeURL)
			require.Nil(t, step.SupportURL)
			require.Nil(t, step.PublishedAt)
			require.Nil(t, step.Source)
			require.Nil(t, step.Deps)
			require.Equal(t, 0, len(step.HostOsTags))
			require.Equal(t, 0, len(step.ProjectTypeTags))
			require.Equal(t, 0, len(step.TypeTags))
			require.Nil(t, step.IsRequiresAdminUser)
			require.Nil(t, step.IsAlwaysRun)
			require.Nil(t, step.IsSkippable)
			require.Nil(t, step.RunIf)
			require.Equal(t, 0, len(step.Inputs))
			require.Equal(t, 0, len(step.Outputs))
		}
	}
}
