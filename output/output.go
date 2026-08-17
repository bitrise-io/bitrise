package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

const (
	// FormatKey ...
	FormatKey = "format"
	// FormatRaw ...
	FormatRaw = "raw"
	// FormatJSON ...
	FormatJSON = "json"
	// FormatYML ...
	FormatYML = "yml"
)

// Format ...
var Format = FormatRaw

// defaultFormat is what ConfigureOutputFormat("") resolves to. It starts as
// FormatRaw so behaviour is unchanged until something calls SetDefault.
var defaultFormat = FormatRaw

// SetDefault sets the format ConfigureOutputFormat falls back to when a
// command's own --format flag is unset. This is what lets a root-persistent
// --output flag (or the "output" config key) take effect without touching
// any per-command --format call site.
func SetDefault(format string) {
	defaultFormat = format
}

// ParseFormat validates a format string without mutating any global state,
// accepting "human" as an alias for FormatRaw.
func ParseFormat(s string) (string, error) {
	switch s {
	case FormatRaw, "human":
		return FormatRaw, nil
	case FormatJSON, FormatYML:
		return s, nil
	default:
		return "", fmt.Errorf("invalid output format: %s", s)
	}
}

// ConfigureOutputFormat sets the global Format from a command's --format flag
// value, falling back to the resolved default (see SetDefault) when empty.
func ConfigureOutputFormat(outFmt string) error {
	if outFmt == "" {
		outFmt = defaultFormat
	}
	parsed, err := ParseFormat(outFmt)
	if err != nil {
		return err
	}
	Format = parsed
	return nil
}

// Render writes result via renderRaw for FormatRaw, or via Print otherwise —
// the "raw table/text vs. json/yml" branch every command needs, in one place.
func Render[T any](w io.Writer, format string, result T, renderRaw func(io.Writer, T) error) error {
	if format == FormatRaw {
		return renderRaw(w, result)
	}
	return Print(w, result, format)
}

// Print marshals outModel per format and writes it to w, indented. Returns an
// error instead of logging it, so a marshaling failure surfaces as a non-zero
// exit rather than a silently successful command.
func Print(w io.Writer, outModel interface{}, format string) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(outModel); err != nil {
			return fmt.Errorf("marshal json output: %w", err)
		}
	case FormatYML:
		serBytes, err := yaml.Marshal(outModel)
		if err != nil {
			return fmt.Errorf("marshal yml output: %w", err)
		}
		if _, err := w.Write(serBytes); err != nil {
			return fmt.Errorf("write yml output: %w", err)
		}
	default:
		return fmt.Errorf("invalid output format: %s", format)
	}
	return nil
}
