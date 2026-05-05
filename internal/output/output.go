// Package output renders command results in human-friendly tables when stdout
// is a TTY, and machine-readable JSON otherwise. The user can override with
// --output {json,yaml,table}.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Format is the chosen output format.
type Format string

const (
	FormatAuto  Format = ""
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
)

// Resolve returns the effective format. If FormatAuto, returns Table when stdout
// is a TTY, JSON otherwise (so output is pipe-friendly by default).
func Resolve(f Format, w io.Writer) Format {
	if f != FormatAuto {
		return f
	}
	if file, ok := w.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		return FormatTable
	}
	return FormatJSON
}

// JSON writes v as pretty-printed JSON to w.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// YAML writes v as YAML to w.
func YAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(v)
}

// Table renders rows as an aligned ASCII table to w.
func Table(w io.Writer, headers []string, rows [][]string) {
	t := tablewriter.NewWriter(w)
	t.Header(headers)
	for _, r := range rows {
		_ = t.Append(r)
	}
	_ = t.Render()
}

// Render auto-dispatches v based on f. For FormatTable the caller must pass a
// rowsFn that converts v into headers + rows; pass nil to fall back to JSON.
func Render(w io.Writer, f Format, v any, headers []string, rowsFn func() [][]string) error {
	switch Resolve(f, w) {
	case FormatJSON:
		return JSON(w, v)
	case FormatYAML:
		return YAML(w, v)
	case FormatTable:
		if rowsFn == nil {
			return JSON(w, v)
		}
		Table(w, headers, rowsFn())
		return nil
	}
	return fmt.Errorf("unknown format: %s", f)
}
