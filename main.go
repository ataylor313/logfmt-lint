// Command logfmt-lint parses logfmt-style log lines (key=value pairs,
// as emitted by Heroku, hashicorp tools, and plenty of Go services),
// checks each line against a small schema, and reports the result as
// either a human-readable summary or JSON.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	jsonOutput := flag.Bool("json", false, "emit a JSON array instead of the human-readable report")
	schemaPath := flag.String("schema", "", "path to a JSON schema file overriding the default required fields, time format, and levels")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [--json] [--schema file] [file ...]\n\nWith no files, reads from stdin.\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	schema := defaultSchema()
	if *schemaPath != "" {
		s, err := loadSchema(*schemaPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		schema = s
	}

	results, err := run(flag.Args(), schema)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *jsonOutput {
		if err := printJSON(os.Stdout, results); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		printHuman(os.Stdout, results)
	}

	for _, r := range results {
		if !r.Valid {
			os.Exit(1)
		}
	}
}

// run reads each named file (or stdin, if paths is empty) and returns a
// Result per non-blank line, in order.
func run(paths []string, schema Schema) ([]Result, error) {
	type source struct {
		name string
		file *os.File
	}

	var sources []source
	if len(paths) == 0 {
		sources = append(sources, source{name: "", file: os.Stdin})
	} else {
		for _, p := range paths {
			f, err := os.Open(p)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			sources = append(sources, source{name: p, file: f})
		}
	}

	var results []Result
	for _, src := range sources {
		scanner := bufio.NewScanner(src.file)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		lineNo := 0
		for scanner.Scan() {
			lineNo++
			raw := scanner.Text()
			if raw == "" {
				continue
			}

			entry, err := parseLine(raw)
			if err != nil {
				results = append(results, Result{
					File:   src.name,
					Line:   lineNo,
					Raw:    raw,
					Valid:  false,
					Errors: []string{err.Error()},
				})
				continue
			}

			errs := validate(entry, schema)
			results = append(results, Result{
				File:   src.name,
				Line:   lineNo,
				Raw:    raw,
				Valid:  len(errs) == 0,
				Errors: errs,
				Fields: fieldsMap(entry.Fields),
			})
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading %s: %w", src.name, err)
		}
	}

	return results, nil
}
