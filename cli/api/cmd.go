package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalapi "github.com/bitrise-io/bitrise/v2/internal/api"
)

// NewCmd returns the `api` command: a generic, curl-like authenticated
// passthrough to the Bitrise API. Unlike every other cloud command, it
// deliberately does not support --format — the response body is always
// printed as-is, so -f is free to mean --field here instead.
func NewCmd() *cobra.Command {
	var method string
	var fields []string
	var headers []string
	var inputPath string
	var all bool
	var include bool

	cmd := &cobra.Command{
		Use:   "api PATH",
		Short: "Make an authenticated request to the Bitrise API",
		Long: `Make an authenticated HTTP request to the Bitrise API and print the response.

PATH is resolved relative to the configured API base URL, or used verbatim
if it's an absolute http(s):// URL. The response body goes to stdout, so it
composes with tools like jq.

--field pairs are sent as strings; use --input for bodies that need nesting
or non-string values.`,
		Example: `  bitrise api /me
  bitrise api /apps -f sort_by=last_build_at --all | jq '.data[].title'
  bitrise api "/apps/APP_ID/builds?limit=10"
  bitrise api /apps/APP_ID/builds -X POST --input body.json
  bitrise api -X DELETE /apps/APP_ID/builds/BUILD_ID -i`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			kvFields, err := parseFields(fields)
			if err != nil {
				return err
			}
			header, err := parseHeaders(headers)
			if err != nil {
				return err
			}

			var body io.Reader
			if inputPath != "" {
				data, err := openInput(cmd.InOrStdin(), inputPath)
				if err != nil {
					return fmt.Errorf("read input: %w", err)
				}
				body = bytes.NewReader(data)
			}

			resolvedMethod := strings.ToUpper(method)
			if resolvedMethod == "" {
				resolvedMethod = http.MethodGet
				if len(kvFields) > 0 || body != nil {
					resolvedMethod = http.MethodPost
				}
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			resp, err := internalapi.NewService(client).Do(cmd.Context(), internalapi.Request{
				Method:   resolvedMethod,
				Path:     args[0],
				Fields:   kvFields,
				Headers:  header,
				Body:     body,
				Paginate: all,
			})
			if err != nil {
				return err
			}

			if err := writeResponse(cmd.OutOrStdout(), resp, include); err != nil {
				return err
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("bitrise API responded with HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "", "HTTP method (default GET, or POST when a body is present)")
	cmd.Flags().StringArrayVarP(&fields, "field", "f", nil, "key=value; query parameter for GET, JSON body field otherwise (may be repeated)")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, `"Name: value"; adds or overrides a request header (may be repeated)`)
	cmd.Flags().StringVar(&inputPath, "input", "", "path to a file for the request body, or - for stdin")
	cmd.Flags().BoolVar(&all, "all", false, "follow cursor pagination and merge every page's data array (GET only)")
	cmd.Flags().BoolVarP(&include, "include", "i", false, "print the response status and headers before the body")
	cmd.MarkFlagsMutuallyExclusive("field", "input")

	return cmd
}

// parseFields parses "key=value" flag values into KeyValue pairs.
func parseFields(raw []string) ([]internalapi.KeyValue, error) {
	fields := make([]internalapi.KeyValue, 0, len(raw))
	for _, f := range raw {
		key, value, ok := strings.Cut(f, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf(`invalid --field %q: expected "key=value"`, f)
		}
		fields = append(fields, internalapi.KeyValue{Key: key, Value: value})
	}
	return fields, nil
}

// parseHeaders parses "Name: value" flag values into an http.Header.
func parseHeaders(raw []string) (http.Header, error) {
	header := http.Header{}
	for _, h := range raw {
		name, value, ok := strings.Cut(h, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, fmt.Errorf(`invalid --header %q: expected "Name: value"`, h)
		}
		header.Add(name, strings.TrimSpace(value))
	}
	return header, nil
}

// openInput reads the request body for --input: stdin when path is "-",
// otherwise the file at path.
func openInput(stdin io.Reader, path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}

// writeResponse prints resp to w: optionally the status line and headers
// (--include), then the body. A JSON body is indented only when w is a real
// terminal; piped or redirected output keeps the API's own formatting so it
// composes with tools like jq. A missing trailing newline is appended either
// way, so a terminal prompt isn't left mid-line.
func writeResponse(w io.Writer, resp internalapi.Response, include bool) error {
	if include {
		statusLine := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if text := http.StatusText(resp.StatusCode); text != "" {
			statusLine += " " + text
		}
		if _, err := fmt.Fprintln(w, statusLine); err != nil {
			return err
		}
		names := make([]string, 0, len(resp.Header))
		for name := range resp.Header {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			for _, value := range resp.Header[name] {
				if _, err := fmt.Fprintf(w, "%s: %s\n", name, value); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	body := resp.Body
	if strings.Contains(resp.Header.Get("Content-Type"), "json") && cmdutil.IsTerminalWriter(w) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, body, "", "  "); err == nil {
			body = pretty.Bytes()
		}
	}

	if _, err := w.Write(body); err != nil {
		return err
	}
	if len(body) > 0 && !bytes.HasSuffix(body, []byte("\n")) {
		_, err := fmt.Fprintln(w)
		return err
	}
	return nil
}
