# Bitrise CLI Trace Logging Example

## Enable Experimental Trace Logging

To enable trace logging in Bitrise CLI, use the `--experimental_trace_logs` flag:

```bash
bitrise run workflow_name --experimental_trace_logs
```

## Example Output

When trace logging is enabled, you'll see structured trace events in the output:

```
BITRISE_TRACE:{"ts":1234567890000000,"type":"workflow_start","step_title":"My Workflow","workflow":"primary","pid":1,"tid":0}
BITRISE_TRACE:{"ts":1234567890001000,"type":"step_start","step_id":"git-clone","step_title":"Git Clone Repository","workflow":"primary","pid":1,"tid":1}
BITRISE_TRACE:{"ts":1234567890020000,"type":"step_end","step_id":"git-clone","step_title":"Git Clone Repository","workflow":"primary","status":"success","duration_us":19000,"pid":1,"tid":1}
BITRISE_TRACE:{"ts":1234567890085000,"type":"workflow_end","step_title":"My Workflow","workflow":"primary","status":"success","duration_us":85000,"pid":1,"tid":0}
```

## Converting to JSON Trace Profile

You can pipe the output to the log-to-json-trace-profile converter:

```bash
bitrise run primary --experimental_trace_logs 2>&1 | ./log-to-json-trace-profile -output trace.json
```

## Viewing in Chrome DevTools

1. Open Chrome DevTools (or go to [text](https://ui.perfetto.dev/))
2. Go to Performance tab
3. Click "Load profile"
4. Select the generated `trace.json` file
5. Analyze the workflow and step performance

## Demo

```shell
# compile bitrise CLI with the new experimental feature
go build -o /tmp/experimental-bitrise-cli

# run `bitrise run` with the new `--experimental_trace_logs` flag
/tmp/experimental-bitrise-cli run trace-log-example-workflow --experimental_trace_logs | tee build-log.txt

# extract and convert the Bitrise Trace Logs into JSON Trace Profile
# NOTE: the log-to-json-trace-profile CLI tool currently lives in a separate private repo
# could be converted to be a Bitrise CLI plugin
log-to-json-trace-profile -input ./build-log.txt -output ./trace.json
```

Additional notes:
- The `log-to-json-trace-profile` could be a Bitrise CLI plugin, to easily extract and convert the timestamps.
- This `log-to-json-trace-profile` based CLI plugin could also be used in Shell scripts: to perform a specific command and print the `BITRISE_TRACE` log traces before and after performing the command, as working with timestamps and micro seconds in Bash is quite tricky.
- The CLI could pass in an env var for the steps, which indicates whether trace logging is enabled or not. The `log-to-json-trace-profile` could get and use it, not printing trace logs when the mode isn't enabled.

Questions/TODO:
- Do we have a way for the step to figure out the `Step ID`?
