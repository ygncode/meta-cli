package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

type Printer struct {
	format Format
	w      io.Writer
}

func New(format Format, w io.Writer) *Printer {
	return &Printer{format: format, w: w}
}

func (p *Printer) Print(v any) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(v)
	case FormatPlain:
		return p.printTSV(v)
	default:
		return p.printTable(v)
	}
}

func (p *Printer) PrintOne(v any) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(v)
	case FormatPlain:
		return p.printTSV(v)
	default:
		return p.printJSON(v)
	}
}

func (p *Printer) OK(msg string) {
	switch p.format {
	case FormatJSON:
		json.NewEncoder(p.w).Encode(map[string]string{"status": "ok", "message": msg})
	default:
		fmt.Fprintln(os.Stderr, msg)
	}
}

func (p *Printer) Err(err error) {
	switch p.format {
	case FormatJSON:
		json.NewEncoder(p.w).Encode(map[string]string{"error": err.Error()})
	default:
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func (p *Printer) printJSON(v any) error {
	enc := json.NewEncoder(p.w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (p *Printer) printTable(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		fmt.Fprintln(p.w, "No results")
		return nil
	}

	headers, rows := extractRows(rv)
	// Truncate long cell values for readable table output
	for i, row := range rows {
		for j, cell := range row {
			rows[i][j] = truncateCell(cell, 60)
		}
	}
	table := tablewriter.NewWriter(p.w)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(rows)
	table.Render()
	return nil
}

func truncateCell(s string, maxLen int) string {
	// Replace newlines with spaces for single-line table display
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ") // collapse multiple spaces
	if len([]rune(s)) > maxLen {
		return string([]rune(s)[:maxLen]) + "..."
	}
	return s
}

func (p *Printer) printTSV(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		return nil
	}

	headers, rows := extractRows(rv)
	w := csv.NewWriter(p.w)
	w.Comma = '\t'
	w.Write(headers)
	for _, row := range rows {
		w.Write(row)
	}
	w.Flush()
	return w.Error()
}

func extractRows(rv reflect.Value) ([]string, [][]string) {
	elemType := rv.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	var headers []string
	var fieldIdxs []int
	for i := 0; i < elemType.NumField(); i++ {
		f := elemType.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			continue
		}
		headers = append(headers, strings.ToUpper(name))
		fieldIdxs = append(fieldIdxs, i)
	}

	var rows [][]string
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		var row []string
		for _, idx := range fieldIdxs {
			row = append(row, fmt.Sprintf("%v", elem.Field(idx).Interface()))
		}
		rows = append(rows, row)
	}
	return headers, rows
}
