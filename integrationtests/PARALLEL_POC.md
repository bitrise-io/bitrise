# Integration Tests Parallelization POC

This document summarizes which integration tests appear safe to run in parallel
across packages (using `go test -count=1`) and which still need to run
sequentially due to shared global state or environment/tooling dependencies.

## Parallel-safe (package-level)

- Single-call `go test ./... -count=1` works when skipping known offenders
  - Verified via `integrationtests/integration_tests_split_run.log`.

## Sequential-only list (run after the parallel pass)

- CLI offline-mode tests (require fastlane on macOS; Stepman route cleanup races):
  - `Test_GivenOfflineMode_WhenStepNotCached_ThenFails`
  - `Test_GivenOnlineMode_WhenStepNotCached_ThenSucceeds`
  - `Test_GivenOfflineMode_WhenStepCached_ThenSuceeds`
- Steps tests that need extra local tooling or flaky steplib usage:
  - `Test_GoModMigration` (needs `node`/npm available)
  - `TestSteplibStepExecutable` (darwin-arm64: no prebuilt binary, step source lacks Go files)
  - `TestNestedStepBundle` (Stepman spec race)
  - `TestStepBundleInputs` (Stepman spec race)
  - `TestStepBundleRunIf` (Stepman spec race)
  - `Test_SensitiveInputs` (Stepman spec race)
- Environment tests that race on Stepman spec updates:
  - `Test_OutputAlias`
  - `Test_EnvstoreTest`
  - `Test_SecretFiltering`
  - `Test_Secret_Filtering_FailingStep`
  - `TestSecretSharing`
- Agent-config tests mutate global config in `$HOME`:
  - `Test_AgentConfigBitriseDirs`
  - `Test_AgentConfigBuildHooksSuccess`
  - `Test_AgentConfigBuildHooksFailure`
- Config test that races Stepman spec reads in parallel:
  - `Test_ModularConfig_Run_JSON_Logs`
  - `Test_ModularConfig`
- CLI tests that sometimes race Stepman spec reads in parallel:
  - `Test_GlobalFlagCI`
  - `Test_JsonParams`
  - `TestConsoleLogCanBeRestoredFromJSONLog`
  - `TestStepDebugLogMessagesAppear`
- CLI test that mutates the CLI binary on disk:
  - `Test_Update`
- Steps test that races Stepman spec reads in parallel:
  - `Test_StepTemplate`
- Workflow test that races Stepman spec reads in parallel:
  - `Test_AsyncStep`
- Workflow test that races Stepman spec reads in parallel:
  - `Test_TimeoutTest`
- Workflow test with before/after envs that flakes under parallel Stepman updates:
  - `Test_WorkflowRunEnvs`

## Sequential-only (known offenders)

- The list above is executed in the sequential pass; all other tests run in
  the parallel pass.

<!-- SEQ_REPORT_START -->
## Where sequential tests concentrate

- Regenerate: `python3 integrationtests/scripts/update_parallel_poc.py`

- By subpackage (count of sequential tests):
  - `cli`: 8
  - `steps`: 7
  - `environment`: 5
  - `config`: 5
  - `workflow`: 3
- By file (count of sequential tests):
  - `integrationtests/config/agent_config_test.go`: 3
  - `integrationtests/cli/offline_mode_test.go`: 3
  - `integrationtests/steps/step_bundles_test.go`: 2
  - `integrationtests/environment/secret_filtering_test.go`: 2
  - `integrationtests/config/modular_config_test.go`: 2
  - `integrationtests/cli/log_format_test.go`: 2
  - `integrationtests/steps/gomodmigrate_test.go`: 1
  - `integrationtests/steps/steplib_step_executable_test.go`: 1
  - `integrationtests/steps/step_bundle_run_if_test.go`: 1
  - `integrationtests/steps/sensitive_inputs_test.go`: 1
  - `integrationtests/environment/output_alias_test.go`: 1
  - `integrationtests/environment/envstore_test.go`: 1
  - `integrationtests/environment/secret_keys_sharing_test.go`: 1
  - `integrationtests/cli/global_flag_test.go`: 1
  - `integrationtests/cli/json_params_test.go`: 1
  - `integrationtests/cli/update_test.go`: 1
  - `integrationtests/steps/step_template_test.go`: 1
  - `integrationtests/workflow/async_step_test.go`: 1
  - `integrationtests/workflow/timeout_test.go`: 1
  - `integrationtests/workflow/workflow_run_envs_test.go`: 1
<!-- SEQ_REPORT_END -->

## Notes

- Initial parallel runs surfaced:
  - agent hook config races
  - Stepman spec races
- Single-call run focuses on parallelism and skips known offenders until they
  are stabilized or gated by environment/tooling availability.
