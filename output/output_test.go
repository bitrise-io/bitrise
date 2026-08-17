package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestParseFormat(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"raw", FormatRaw, false},
		{"human", FormatRaw, false}, // alias for raw
		{"json", FormatJSON, false},
		{"yml", FormatYML, false},
		{"yaml", "", true}, // not an accepted spelling
		{"", "", true},
		{"RAW", "", true}, // case-sensitive on purpose
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseFormat(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ParseFormat(%q): expected error, got %q", c.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFormat(%q): unexpected error: %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("ParseFormat(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestConfigureOutputFormat_FallsBackToDefault(t *testing.T) {
	t.Cleanup(func() { SetDefault(FormatRaw) })

	SetDefault(FormatJSON)
	if err := ConfigureOutputFormat(""); err != nil {
		t.Fatalf("ConfigureOutputFormat(\"\"): unexpected error: %v", err)
	}
	if Format != FormatJSON {
		t.Fatalf("Format = %q, want %q", Format, FormatJSON)
	}
}

func TestConfigureOutputFormat_ExplicitOverridesDefault(t *testing.T) {
	t.Cleanup(func() { SetDefault(FormatRaw) })

	SetDefault(FormatJSON)
	if err := ConfigureOutputFormat(FormatYML); err != nil {
		t.Fatalf("ConfigureOutputFormat(%q): unexpected error: %v", FormatYML, err)
	}
	if Format != FormatYML {
		t.Fatalf("Format = %q, want %q", Format, FormatYML)
	}
}

func TestConfigureOutputFormat_InvalidFormat(t *testing.T) {
	if err := ConfigureOutputFormat("bogus"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

type sample struct {
	Name string `json:"name"`
	N    int    `json:"n"`
}

func TestPrint_JSONIsIndentedAndWritesToW(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, sample{Name: "x", N: 7}, FormatJSON); err != nil {
		t.Fatalf("Print JSON: %v", err)
	}
	var got sample
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %q", err, buf.String())
	}
	if got != (sample{Name: "x", N: 7}) {
		t.Fatalf("decoded JSON differs: got %+v", got)
	}
	if !strings.Contains(buf.String(), "\n  ") {
		t.Errorf("expected indented JSON, got %q", buf.String())
	}
}

func TestPrint_YMLWritesToW(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, sample{Name: "widget", N: 3}, FormatYML); err != nil {
		t.Fatalf("Print YML: %v", err)
	}
	if !strings.Contains(buf.String(), "name: widget") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPrint_InvalidFormat(t *testing.T) {
	if err := Print(io.Discard, sample{}, "bogus"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestRender_Raw(t *testing.T) {
	var buf bytes.Buffer
	renderRaw := func(w io.Writer, v sample) error {
		_, err := w.Write([]byte("name=" + v.Name))
		return err
	}
	if err := Render(&buf, FormatRaw, sample{Name: "z"}, renderRaw); err != nil {
		t.Fatalf("Render raw: %v", err)
	}
	if buf.String() != "name=z" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestRender_JSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Render[sample](&buf, FormatJSON, sample{Name: "x", N: 7}, nil); err != nil {
		t.Fatalf("Render JSON: %v", err)
	}
	var got sample
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %q", err, buf.String())
	}
	if got != (sample{Name: "x", N: 7}) {
		t.Fatalf("decoded JSON differs: got %+v", got)
	}
}

func TestRender_PropagatesRawError(t *testing.T) {
	wantErr := errors.New("boom")
	renderRaw := func(io.Writer, sample) error { return wantErr }
	err := Render(io.Discard, FormatRaw, sample{}, renderRaw)
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}
