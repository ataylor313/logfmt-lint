package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// Result is one line's worth of parse + validation output. It is also the
// unit that gets marshaled for --json, so its field names and omitempty
// tags are the actual JSON contract, not incidental.
type Result struct {
	File   string            `json:"file,omitempty"`
	Line   int               `json:"line"`
	Raw    string            `json:"raw"`
	Valid  bool              `json:"valid"`
	Errors []string          `json:"errors,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

func fieldsMap(fields []Field) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	m := make(map[string]string, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}

func printHuman(w io.Writer, results []Result) {
	valid := 0
	for _, r := range results {
		loc := fmt.Sprintf("line %d", r.Line)
		if r.File != "" {
			loc = fmt.Sprintf("%s:%d", r.File, r.Line)
		}
		if r.Valid {
			valid++
			fmt.Fprintf(w, "%s: ok\n", loc)
			continue
		}
		fmt.Fprintf(w, "%s: invalid\n", loc)
		for _, e := range r.Errors {
			fmt.Fprintf(w, "    %s\n", e)
		}
		fmt.Fprintf(w, "    raw: %s\n", r.Raw)
	}
	fmt.Fprintf(w, "\n%d lines, %d valid, %d invalid\n", len(results), valid, len(results)-valid)
}

func printJSON(w io.Writer, results []Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
