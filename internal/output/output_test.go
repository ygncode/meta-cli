package output_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/output"
)

type testItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatJSON, &buf)

	items := []testItem{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
	}

	if err := p.Print(items); err != nil {
		t.Fatalf("Print: %v", err)
	}

	var got []testItem
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 items, got %d", len(got))
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatTable, &buf)

	items := []testItem{
		{ID: "1", Name: "Alice"},
	}

	if err := p.Print(items); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Errorf("expected headers in table output, got:\n%s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected data in table output, got:\n%s", out)
	}
}

func TestPrintTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatTable, &buf)

	if err := p.Print([]testItem{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	if !strings.Contains(buf.String(), "No results") {
		t.Errorf("expected 'No results' for empty slice, got: %s", buf.String())
	}
}

func TestPrintPlain(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatPlain, &buf)

	items := []testItem{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
	}

	if err := p.Print(items); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("expected 3 lines (header+2), got %d: %q", len(lines), out)
	}
	// Verify tab-separated
	if !strings.Contains(lines[0], "\t") {
		t.Errorf("expected tab-separated header, got: %s", lines[0])
	}
}

func TestPrintPlainEmpty(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatPlain, &buf)

	if err := p.Print([]testItem{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output for empty slice, got: %q", buf.String())
	}
}

func TestPrintOne(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatJSON, &buf)

	item := testItem{ID: "1", Name: "Alice"}
	if err := p.PrintOne(item); err != nil {
		t.Fatalf("PrintOne: %v", err)
	}

	var got testItem
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "1" {
		t.Errorf("expected ID 1, got %s", got.ID)
	}
}

func TestOKJSON(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatJSON, &buf)
	p.OK("done")

	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf("expected status ok, got %v", got)
	}
}

func TestErrJSON(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatJSON, &buf)
	p.Err(errors.New("something failed"))

	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["error"] != "something failed" {
		t.Errorf("expected error message, got %v", got)
	}
}

func TestPrintNonSlice(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatTable, &buf)

	// Passing a non-slice should print "No results"
	if err := p.Print("not a slice"); err != nil {
		t.Fatalf("Print: %v", err)
	}
	if !strings.Contains(buf.String(), "No results") {
		t.Errorf("expected 'No results' for non-slice, got: %s", buf.String())
	}
}

type itemWithOmitted struct {
	ID     string `json:"id"`
	Hidden string `json:"-"`
}

func TestPrintOmitsHiddenFields(t *testing.T) {
	var buf bytes.Buffer
	p := output.New(output.FormatTable, &buf)

	items := []itemWithOmitted{{ID: "1", Hidden: "secret"}}
	if err := p.Print(items); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "secret") {
		t.Errorf("expected hidden field to be omitted, got: %s", out)
	}
}
