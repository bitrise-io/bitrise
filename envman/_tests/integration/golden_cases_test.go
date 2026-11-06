package integration

import "strings"

const (
	kb             = 1024
	defaultLimitKB = 256
	// The release that added the current ENVMAN_ENV_BYTES_LIMIT_IN_KB and ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB
	// overrides (envman#253). 2.6.0 had an earlier version that 2.6.1 dropped, and the 2.5.5 the stacks ship
	// ignores them.
	envLimitOverridesVersion = "2.7.0"
)

const printEnvs = `printf 'A=%s|B=%s|C=%s\n' "${A-<unset>}" "${B-<unset>}" "${C-<unset>}"`

// goldenCases are the inputs of TestGolden and the only part of the suite maintained by hand. Each case
// runs its steps in a fresh work dir, and its expected output lives in testdata/generated/<name>.yml.
//
// To add or change a case, edit it here, run `make envman-golden` and review the diff of the generated
// file. Never edit the generated files by hand.
//
// A case with minVersion records behavior that first shipped in that standalone envman release, so
// older binaries skip it. `make envman-golden` uses the in-repo build, which writes these cases too.
var goldenCases = []goldenCase{
	// --- global flags, help and version
	{name: "help_no_args", steps: []goldenStep{envmanCmd()}},
	{name: "help_flag", steps: []goldenStep{envmanCmd("--help"), envmanCmd("-h")}},
	{name: "help_command", steps: []goldenStep{envmanCmd("help"), envmanCmd("help", "add"), envmanCmd("h")}},
	{name: "help_subcommands", steps: initThen(
		envmanCmd("version", "-h"),
		envmanCmd("init", "-h"),
		envmanCmd("add", "--help"),
		envmanCmd("clear", "-h"),
		envmanCmd("print", "-h"),
		envmanCmd("run", "-h"),
		envmanCmd("unset", "-h"),
	)},
	{name: "version", steps: []goldenStep{
		envmanCmd("--version"),
		envmanCmd("-v"),
		envmanCmd("version"),
		envmanCmd("version", "--format", "json"),
		envmanCmd("version", "--format", "yml"),
		envmanCmd("version", "--format", "raw"),
		envmanCmd("version", "--format", "invalid"),
	}},
	{name: "unknown_command", steps: []goldenStep{envmanCmd("nope")}},
	{name: "unknown_global_flag", steps: []goldenStep{envmanCmd("--nope", "init")}},
	{name: "unknown_command_flag", steps: []goldenStep{envmanCmd("add", "--key", "A", "--nope")}},
	{name: "loglevel", steps: []goldenStep{
		envmanCmd("-l", "debug", "init"),
		envmanCmd("--loglevel", "debug", "add", "--key", "A", "--value", "a"),
		envmanCmd("init").withEnv("LOGLEVEL=debug"),
		envmanCmd("-l", "warn", "init").withEnv("LOGLEVEL=debug"),
		envmanCmd("-l", "invalid", "init"),
	}},
	{name: "tool_mode", steps: []goldenStep{
		envmanCmd("-t", "init"),
		envmanCmd("--tool", "init"),
		envmanCmd("init").withEnv("ENVMAN_TOOLMODE=true"),
		envmanCmd("init").withEnv("ENVMAN_TOOLMODE=invalid"),
	}},
	{name: "path", steps: []goldenStep{
		envmanCmd("-p", "$WORK/short.yml", "init"),
		envmanCmd("-p", "$WORK/short.yml", "add", "--key", "A", "--value", "short"),
		envmanCmd("--path", "$WORK/long.yml", "init"),
		envmanCmd("--path", "$WORK/long.yml", "add", "--key", "A", "--value", "long"),
		envmanCmd("init").withEnv("ENVMAN_ENVSTORE_PATH=$WORK/env.yml"),
		envmanCmd("add", "--key", "A", "--value", "env").withEnv("ENVMAN_ENVSTORE_PATH=$WORK/env.yml"),
		envmanCmd("-p", "$WORK/short.yml", "add", "--key", "A", "--value", "flag").withEnv("ENVMAN_ENVSTORE_PATH=$WORK/env.yml"),
		envmanCmd("-p", "relative.yml", "init"),
		envmanCmd("-p", "relative.yml", "add", "--key", "A", "--value", "relative"),
		envmanCmd("-p", "$WORK/missing-dir/store.yml", "init"),
		envmanCmd("-p", "", "init"),
	}},

	// --- init / clear
	{name: "init", steps: []goldenStep{
		envmanCmd("init"),
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("init"),
		envmanCmd("init", "--clear"),
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("i", "-c"),
	}},
	{name: "clear", steps: []goldenStep{
		envmanCmd("clear"),
		envmanCmd("init"),
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("clear"),
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("c"),
	}},

	// --- add
	{name: "add_flags", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "long"),
		envmanCmd("add", "-k", "B", "-v", "short"),
		envmanCmd("a", "-k", "C", "-v", "alias"),
		envmanCmd("add", "--key=D", "--value=equals"),
	)},
	{name: "add_replace_and_append", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "1"),
		envmanCmd("add", "--key", "A", "--value", "2"),
		envmanCmd("add", "--key", "A", "--value", "3", "--append"),
		envmanCmd("add", "--key", "A", "--value", "4", "-a"),
		envmanCmd("-t", "add", "--key", "A", "--value", "5"),
		envmanCmd("add", "--key", "A", "--value", "6"),
		envmanCmd("add", "--key", "A", "--value", "7").withStdin("replace\n"),
	)},
	{name: "add_missing_key", steps: initThen(
		envmanCmd("add", "--value", "a"),
		envmanCmd("add", "--key", "", "--value", "a"),
	)},
	{name: "add_empty_value", steps: initThen(
		envmanCmd("add", "--key", "A"),
		envmanCmd("add", "--key", "B", "--value", ""),
		envmanCmd("add", "--key", "C").withStdin(""),
	)},
	{name: "add_value_sources", files: map[string]string{
		"value.txt":   "from file\n",
		"empty.txt":   "",
		"unicode.txt": "árvíztűrő tükörfúrógép 🚀",
	}, steps: initThen(
		envmanCmd("add", "--key", "FILE", "--valuefile", "value.txt"),
		envmanCmd("add", "--key", "FILE_SHORT", "-f", "$WORK/unicode.txt"),
		envmanCmd("add", "--key", "PIPE").withStdin("from stdin"),
		envmanCmd("add", "--key", "FLAG_OVER_FILE", "--value", "flag", "--valuefile", "value.txt"),
		envmanCmd("add", "--key", "FLAG_OVER_PIPE", "--value", "flag").withStdin("stdin"),
		envmanCmd("add", "--key", "EMPTY_FLAG_FALLS_TO_FILE", "--value", "", "--valuefile", "value.txt"),
		envmanCmd("add", "--key", "EMPTY_FILE_FALLS_TO_PIPE", "--valuefile", "empty.txt").withStdin("stdin"),
		envmanCmd("add", "--key", "FILE_OVER_PIPE", "--valuefile", "value.txt").withStdin("stdin"),
		envmanCmd("add", "--key", "MISSING_FILE", "--valuefile", "missing.txt"),
	)},
	{name: "add_options", steps: initThen(
		envmanCmd("add", "--key", "NO_EXPAND", "--value", "$HOME", "--no-expand"),
		envmanCmd("add", "--key", "NO_EXPAND_SHORT", "--value", "$HOME", "-n"),
		envmanCmd("add", "--key", "SKIP", "--value", "", "--skip-if-empty"),
		envmanCmd("add", "--key", "SENSITIVE", "--value", "secret", "--sensitive"),
		envmanCmd("add", "--key", "ALL", "--value", "x", "-n", "--skip-if-empty", "--sensitive", "-a"),
	)},
	{name: "add_value_formats", steps: initThen(
		envmanCmd("add", "--key", "COLON", "--value", "a: b"),
		envmanCmd("add", "--key", "LEADING_SPACE", "--value", "  lead"),
		envmanCmd("add", "--key", "TRAILING_SPACE", "--value", "trail  "),
		envmanCmd("add", "--key", "HASH", "--value", "# not a comment"),
		envmanCmd("add", "--key", "BOOL", "--value", "true"),
		envmanCmd("add", "--key", "NUMBER", "--value", "0123"),
		envmanCmd("add", "--key", "NULL", "--value", "null"),
		envmanCmd("add", "--key", "TILDE", "--value", "~"),
		envmanCmd("add", "--key", "QUOTES", "--value", `"double" 'single'`),
		envmanCmd("add", "--key", "MULTILINE").withStdin("line1\nline2\n\nline4\n"),
		envmanCmd("add", "--key", "CRLF").withStdin("a\r\nb"),
		envmanCmd("add", "--key", "TAB").withStdin("a\tb"),
		envmanCmd("add", "--key", "UNICODE", "--value", "ünï©ødé ✓"),
		envmanCmd("add", "--key", "YAML_LIKE", "--value", "- [a, {b: c}]"),
		envmanCmd("add", "--key", "lower.dotted-key", "--value", "k"),
		envmanCmd("print", "--format", "json"),
	)},
	{name: "missing_store", steps: []goldenStep{
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("unset", "--key", "A"),
		envmanCmd("print"),
		envmanCmd("clear"),
		envmanCmd("run", "true"),
	}},
	{name: "add_debug_log", steps: initThen(
		envmanCmd("-l", "debug", "add", "--key", "A", "--value", "a"),
		envmanCmd("-l", "debug", "add", "--key", "B", "--value", "$A", "--no-expand"),
		envmanCmd("-l", "debug", "add", "--key", "C").withStdin("piped"),
	)},

	// --- envstore file format (reading hand written / legacy stores)
	{name: "store_formats", files: map[string]string{
		"empty.yml":      "",
		"empty_list.yml": "envs: []\n",
		"no_envs.yml":    "other: value\n",
		"typed.yml":      "envs:\n- INT: 1\n- FLOAT: 1.5\n- BOOL: true\n- NULL_VALUE: null\n- QUOTED: \"x\"\n",
		"null_key.yml":   "envs:\n- NULL: a\n",
		"opts.yml": `envs:
- A: a
  opts:
    title: Title
    description: Description
    summary: Summary
    category: Category
    value_options: ["a", "b"]
    is_required: true
    is_expand: false
    is_dont_change_value: true
    is_template: true
    is_sensitive: true
    skip_if_empty: true
    meta:
      key: value
- B: $A
`,
		"invalid.yml":     "envs: [\n",
		"invalid_env.yml": "envs:\n- A: a\n  B: b\n",
		"not_a_list.yml":  "envs: value\n",
	}, steps: []goldenStep{
		envmanCmd("-p", "empty.yml", "print", "--format", "json"),
		envmanCmd("-p", "empty_list.yml", "print", "--format", "json"),
		envmanCmd("-p", "no_envs.yml", "print", "--format", "json"),
		envmanCmd("-p", "typed.yml", "print", "--format", "json"),
		envmanCmd("-p", "typed.yml", "run", "sh", "-c", `printf '%s|%s|%s|%s|%s\n' "$INT" "$FLOAT" "$BOOL" "${NULL_VALUE-<unset>}" "$QUOTED"`),
		envmanCmd("-p", "opts.yml", "print", "--format", "json", "--expand"),
		envmanCmd("-p", "opts.yml", "add", "--key", "C", "--value", "c"),
		envmanCmd("-p", "null_key.yml", "print", "--format", "json"),
		envmanCmd("-p", "invalid.yml", "print"),
		envmanCmd("-p", "invalid.yml", "add", "--key", "A", "--value", "a"),
		envmanCmd("-p", "invalid_env.yml", "print"),
		envmanCmd("-p", "not_a_list.yml", "print"),
		envmanCmd("-p", "missing.yml", "print"),
		envmanCmd("-p", "missing.yml", "run", "true"),
	}},
	{name: "store_roundtrip", files: map[string]string{
		".envstore.yml": "envs:\n- A: a\n- A: duplicate\n- B: \"quoted\"\n  opts:\n    is_expand: true\n    skip_if_empty: false\n- C: c # comment\n",
	}, steps: []goldenStep{
		envmanCmd("print", "--format", "json"),
		envmanCmd("add", "--key", "D", "--value", "d"),
	}},
	{name: "empty_store_file", files: map[string]string{".envstore.yml": ""}, env: []string{"DONOTEXPORT=os", "EXPORT_THIS=os"}, steps: []goldenStep{
		envmanCmd("add", "--key", "KEY", "--value", "value"),
		envmanCmd("add", "--key", "ADDED", "--value", "sample value"),
		envmanCmd("unset", "--key", "ADDED"),
		envmanCmd("unset", "--key", "DONOTEXPORT"),
		envmanCmd("run", "sh", "-c", `printf '%s|%s|%s|%s\n' "${KEY-<unset>}" "${ADDED-<unset>}" "${DONOTEXPORT-<unset>}" "${EXPORT_THIS-<unset>}"`),
	}},

	// --- value size limit
	{name: "value_limit_default", files: map[string]string{
		"at_limit":   value(defaultLimitKB * kb),
		"over_limit": value(defaultLimitKB*kb + 1),
		"multibyte":  strings.Repeat("é", defaultLimitKB*kb/2) + "a",
	}, steps: initThen(
		envmanCmd("add", "--key", "OVER", "--valuefile", "over_limit"),
		envmanCmd("add", "--key", "MULTIBYTE", "--valuefile", "multibyte"),
		envmanCmd("add", "--key", "AT", "--valuefile", "at_limit"),
	)},
	{name: "value_limit_env_override", minVersion: envLimitOverridesVersion, files: map[string]string{
		"1kb":      value(kb),
		"1kb_plus": value(kb + 1),
		"2kb":      value(2 * kb),
	}, steps: initThen(
		envmanCmd("add", "--key", "OVER", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=1"),
		envmanCmd("add", "--key", "AT", "--valuefile", "1kb").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=1"),
		envmanCmd("add", "--key", "RAISED", "--valuefile", "2kb").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=2"),
	)},
	{name: "value_limit_config_file", minVersion: envLimitOverridesVersion, files: map[string]string{
		"home/.envman/configs.json": `{"env_bytes_limit_in_kb": 1}`,
		"1kb":                       value(kb),
		"1kb_plus":                  value(kb + 1),
	}, steps: initThen(
		envmanCmd("add", "--key", "OVER", "--valuefile", "1kb_plus"),
		envmanCmd("add", "--key", "AT", "--valuefile", "1kb"),
		envmanCmd("add", "--key", "ENV_WINS", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=2"),
		envmanCmd("add", "--key", "INVALID_ENV", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=abc"),
		envmanCmd("add", "--key", "NEGATIVE_ENV", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=-1"),
		envmanCmd("add", "--key", "EMPTY_ENV", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB="),
		envmanCmd("add", "--key", "FLOAT_ENV", "--valuefile", "1kb_plus").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=1.5"),
	)},
	{name: "value_limit_zero_is_unlimited", minVersion: envLimitOverridesVersion, files: map[string]string{
		"big": value((defaultLimitKB + 1) * kb),
	}, steps: initThen(
		envmanCmd("add", "--key", "ONLY_VALUE_UNLIMITED", "--valuefile", "big").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=0"),
		envmanCmd("add", "--key", "BOTH_UNLIMITED", "--valuefile", "big").withEnv("ENVMAN_ENV_BYTES_LIMIT_IN_KB=0", "ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB=0"),
	)},

	// --- list size limit
	{name: "list_limit_default", files: map[string]string{
		"half":     value(defaultLimitKB * kb / 2),
		"one_byte": "a",
	}, steps: initThen(
		envmanCmd("add", "--key", "A", "--valuefile", "half"),
		envmanCmd("add", "--key", "B", "--valuefile", "half"),
		envmanCmd("add", "--key", "C", "--valuefile", "one_byte"),
		envmanCmd("add", "--key", "EMPTY", "--value", ""),
	)},
	{name: "list_limit_env_override", minVersion: envLimitOverridesVersion, env: []string{"ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB=1"}, files: map[string]string{
		"600": value(600),
		"424": value(424),
		"100": value(100),
	}, steps: initThen(
		envmanCmd("add", "--key", "A", "--valuefile", "600"),
		envmanCmd("add", "--key", "B", "--valuefile", "424"),
		envmanCmd("add", "--key", "C", "--value", "x"),
		envmanCmd("add", "--key", "A", "--valuefile", "100"),
		envmanCmd("add", "--key", "LONG_KEY_NAME_IS_NOT_COUNTED_IN_THE_LIST_SIZE", "--value", ""),
	)},
	{name: "list_limit_config_file", minVersion: envLimitOverridesVersion, files: map[string]string{
		"home/.envman/configs.json": `{"env_list_bytes_limit_in_kb": 1}`,
		"1kb":                       value(kb),
	}, steps: initThen(
		envmanCmd("add", "--key", "A", "--valuefile", "1kb"),
		envmanCmd("add", "--key", "B", "--value", "x"),
		envmanCmd("add", "--key", "B", "--value", "x").withEnv("ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB=2"),
		envmanCmd("add", "--key", "C", "--value", "x").withEnv("ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB=invalid"),
	)},
	{name: "list_limit_counts_existing_store", minVersion: envLimitOverridesVersion, files: map[string]string{
		".envstore.yml": "envs:\n- A: " + value(kb) + "\n- B: $A\n  opts:\n    is_expand: true\n",
	}, env: []string{"ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB=1"}, steps: []goldenStep{
		envmanCmd("add", "--key", "C", "--value", "x"),
		envmanCmd("unset", "--key", "C"),
	}},

	// --- configs.json
	{name: "config_file_both_limits", files: map[string]string{
		"home/.envman/configs.json": `{"env_bytes_limit_in_kb": 2, "env_list_bytes_limit_in_kb": 3}`,
		"2kb":                       value(2 * kb),
		"2kb_plus":                  value(2*kb + 1),
	}, steps: initThen(
		envmanCmd("add", "--key", "OVER", "--valuefile", "2kb_plus"),
		envmanCmd("add", "--key", "A", "--valuefile", "2kb"),
		envmanCmd("add", "--key", "B", "--valuefile", "2kb"),
	)},
	{name: "config_file_zero_and_negative", files: map[string]string{
		"home/.envman/configs.json": `{"env_bytes_limit_in_kb": 0, "env_list_bytes_limit_in_kb": -1}`,
		"big":                       value((defaultLimitKB + 1) * kb),
	}, steps: initThen(
		envmanCmd("add", "--key", "A", "--valuefile", "big"),
	)},
	{name: "config_file_malformed", files: map[string]string{
		"home/.envman/configs.json": `{"env_bytes_limit_in_kb": `,
	}, steps: []goldenStep{
		envmanCmd("init"),
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("print"),
		envmanCmd("run", "true"),
		envmanCmd("version"),
		envmanCmd("--version"),
	}},
	{name: "config_file_wrong_types", files: map[string]string{
		"home/.envman/configs.json": `{"env_bytes_limit_in_kb": "1"}`,
	}, steps: []goldenStep{
		envmanCmd("add", "--key", "A", "--value", "a"),
	}},
	{name: "config_file_unknown_keys_and_empty", files: map[string]string{
		"home/.envman/configs.json":  `{"other": 1}`,
		"home2/.envman/configs.json": ``,
	}, steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("add", "--key", "B", "--value", "b").withEnv("HOME=$WORK/home2"),
	)},

	// --- expansion
	{name: "expand_references", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("add", "--key", "B", "--value", "$A-${A}-$$A-\\$A"),
		envmanCmd("add", "--key", "C", "--value", "$B/$UNDEFINED/${UNDEFINED}/end"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "expand_no_expand", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("add", "--key", "B", "--value", "$A", "--no-expand"),
		envmanCmd("add", "--key", "C", "--value", "$B"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "expand_order_and_self_reference", env: []string{"A=os"}, steps: initThen(
		envmanCmd("add", "--key", "B", "--value", "$A"),
		envmanCmd("add", "--key", "A", "--value", "$A:store"),
		envmanCmd("add", "--key", "C", "--value", "$A"),
		envmanCmd("add", "--key", "A", "--value", "$A:again", "--append"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "expand_os_env", env: []string{"A=from-os", "B=os-b"}, steps: initThen(
		envmanCmd("add", "--key", "C", "--value", "$A+$B"),
		envmanCmd("add", "--key", "B", "--value", "store-b"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "expand_skip_if_empty", env: []string{"A=os"}, steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "", "--skip-if-empty"),
		envmanCmd("add", "--key", "B", "--value", "$A"),
		envmanCmd("add", "--key", "C", "--value", "$EMPTY", "--skip-if-empty"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "expand_unset", env: []string{"A=os", "B=os"}, steps: initThen(
		envmanCmd("unset", "--key", "A"),
		envmanCmd("add", "--key", "C", "--value", "${A}x"),
		envmanCmd("rm", "--key", "B"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
		envmanCmd("add", "--key", "A", "--value", "readded"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},

	// --- sensitive values
	{name: "sensitive", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "public"),
		envmanCmd("add", "--key", "B", "--value", "secret", "--sensitive"),
		envmanCmd("add", "--key", "C", "--value", "$B-derived", "--sensitive"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "json", "--sensitive-only"),
		envmanCmd("print", "--format", "json", "--sensitive-only", "--expand"),
		envmanCmd("print", "--sensitive-only"),
		envmanCmd("-l", "debug", "add", "--key", "D", "--value", "debug-secret", "--sensitive"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},
	{name: "sensitive_only_empty", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "public"),
		envmanCmd("print", "--format", "json", "--sensitive-only"),
	)},

	// --- print
	{name: "print_formats", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "a"),
		envmanCmd("add", "--key", "B", "--value", "multi\nline \"quoted\" $A"),
		envmanCmd("print"),
		envmanCmd("print", "--format", "raw"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "envlist"),
		envmanCmd("print", "--format", "envlist", "--expand"),
		envmanCmd("p", "--format=json"),
		envmanCmd("print", "--format", "invalid"),
		envmanCmd("print", "--format", ""),
	)},
	{name: "print_empty_store", steps: []goldenStep{
		envmanCmd("init"),
		envmanCmd("print"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "envlist"),
	}},
	{name: "print_duplicate_keys", steps: initThen(
		envmanCmd("add", "--key", "A", "--value", "first"),
		envmanCmd("add", "--key", "A", "--value", "second", "--append"),
		envmanCmd("print", "--format", "json"),
		envmanCmd("print", "--format", "json", "--expand"),
		envmanCmd("run", "sh", "-c", printEnvs),
	)},

	// --- run
	{name: "run_exit_codes", steps: initThen(
		envmanCmd("run", "sh", "-c", "printf out; printf err >&2"),
		envmanCmd("run", "sh", "-c", "exit 1"),
		envmanCmd("run", "sh", "-c", "exit 42"),
		envmanCmd("run", "sh", "-c", "kill -TERM $$"),
		envmanCmd("run", "envman-golden-command-not-found"),
		envmanCmd("run"),
		envmanCmd("r", "sh", "-c", "exit 3"),
	)},
	{name: "run_args_and_stdin", steps: initThen(
		envmanCmd("run", "sh", "-c", `printf '[%s]' "$@"`, "argv0", "--key", "-h", "--version", "a b"),
		envmanCmd("run", "--help"),
		envmanCmd("run", "cat").withStdin("from stdin\n"),
		envmanCmd("run", "sh", "-c", `printf '%s' "$PWD"`),
		envmanCmd("-p", "$WORK/custom.yml", "init"),
		envmanCmd("-p", "$WORK/custom.yml", "run", "sh", "-c", `printf '%s' "${ENVMAN_ENVSTORE_PATH-<unset>}"`),
	)},
	{name: "run_env_inheritance", env: []string{"OS_ONLY=os", "OVERRIDDEN=os"}, steps: initThen(
		envmanCmd("add", "--key", "OVERRIDDEN", "--value", "store"),
		envmanCmd("run", "sh", "-c", `printf '%s|%s|%s' "$OS_ONLY" "$OVERRIDDEN" "$HOME"`),
	)},

	// --- unset
	{name: "unset", env: []string{"A=os"}, steps: initThen(
		envmanCmd("unset", "--key", "A"),
		envmanCmd("unset", "--key", "MISSING"),
		envmanCmd("add", "--key", "B", "--value", "b"),
		envmanCmd("unset", "--key", "B"),
		envmanCmd("rm", "--key", "C"),
		envmanCmd("unset"),
		envmanCmd("unset", "--key", ""),
	)},
}

func value(size int) string {
	return strings.Repeat("a", size)
}

func envmanCmd(args ...string) goldenStep {
	return goldenStep{args: args}
}

func (s goldenStep) withStdin(stdin string) goldenStep {
	s.stdin = &stdin
	return s
}

func (s goldenStep) withEnv(env ...string) goldenStep {
	s.env = append(s.env, env...)
	return s
}

func initThen(steps ...goldenStep) []goldenStep {
	return append([]goldenStep{envmanCmd("init")}, steps...)
}
