package main

import (
	"fmt"
	"strconv"
)

// Field is one key=value pair in the order it appeared on the line.
type Field struct {
	Key   string
	Value string
}

// Entry is a single parsed log line.
type Entry struct {
	Raw    string
	Fields []Field
}

// Get returns the value of the first field with the given key.
func (e Entry) Get(key string) (string, bool) {
	for _, f := range e.Fields {
		if f.Key == key {
			return f.Value, true
		}
	}
	return "", false
}

// parseLine parses one logfmt line: space-separated key=value pairs where
// the value may be a bare token or a double-quoted string. Quoted values
// are unquoted with strconv.Unquote, so they follow Go escaping rules
// (\", \\, \n, ...) rather than a bespoke escaping scheme.
func parseLine(raw string) (Entry, error) {
	var fields []Field
	i, n := 0, len(raw)

	for i < n {
		for i < n && raw[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		start := i
		for i < n && raw[i] != '=' && raw[i] != ' ' {
			i++
		}
		if i >= n || raw[i] != '=' {
			return Entry{}, fmt.Errorf("malformed token %q: missing '='", raw[start:i])
		}
		key := raw[start:i]
		if key == "" {
			return Entry{}, fmt.Errorf("empty key at byte %d", start)
		}
		i++ // consume '='

		var value string
		if i < n && raw[i] == '"' {
			qstart := i
			i++
			for i < n && raw[i] != '"' {
				if raw[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			if i >= n {
				return Entry{}, fmt.Errorf("unterminated quoted value for key %q", key)
			}
			i++ // consume closing quote
			unquoted, err := strconv.Unquote(raw[qstart:i])
			if err != nil {
				return Entry{}, fmt.Errorf("invalid quoted value for key %q: %w", key, err)
			}
			value = unquoted
		} else {
			vstart := i
			for i < n && raw[i] != ' ' {
				i++
			}
			value = raw[vstart:i]
		}

		fields = append(fields, Field{Key: key, Value: value})
	}

	return Entry{Raw: raw, Fields: fields}, nil
}
