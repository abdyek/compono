package compono

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// formatDiagnostics formats diagnostics for .diag golden files, one per line:
// "<code> <source> <start line>:<start column>-<end line>:<end column> <calls>".
// The source is the scope path joined with "/", or "-" for the converted
// source. The calls are "<kind>:<name>" joined with ">", or "-" if there are
// none.
func formatDiagnostics(diags []Diagnostic) string {
	lines := make([]string, 0, len(diags))
	for _, d := range diags {
		source := strings.Join(d.Source, "/")
		if source == "" {
			source = "-"
		}

		var calls string
		if len(d.Calls) == 0 {
			calls = "-"
		} else {
			parts := make([]string, 0, len(d.Calls))
			for _, c := range d.Calls {
				parts = append(parts, fmt.Sprintf("%s:%s", c.Kind, c.Name))
			}
			calls = strings.Join(parts, ">")
		}

		lines = append(lines, fmt.Sprintf(
			"%s %s %d:%d-%d:%d %s",
			d.Code,
			source,
			d.Range.Start.Line,
			d.Range.Start.Column,
			d.Range.End.Line,
			d.Range.End.Column,
			calls,
		))
	}
	return strings.Join(lines, "\n")
}

// readDiagnosticsGolden returns the trimmed content of a .diag golden file, or
// "" if the file does not exist.
func readDiagnosticsGolden(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ""
		}
		t.Fatalf("read diagnostics golden %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}
