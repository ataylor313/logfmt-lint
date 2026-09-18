package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Schema describes which fields a line must carry and how two of them
// (a timestamp and a level) are checked beyond mere presence. It's the
// configurable replacement for the values that used to be hardcoded in
// validate.go.
type Schema struct {
	Required   []string `json:"required"`
	TimeField  string   `json:"time_field"`
	TimeFormat string   `json:"time_format"`
	LevelField string   `json:"level_field"`
	Levels     []string `json:"levels"`
}

// defaultSchema reproduces the behavior logfmt-lint has always had: time
// must be RFC3339, level must be one of the five syslog-ish values, and
// msg just has to be present.
func defaultSchema() Schema {
	return Schema{
		Required:   []string{"time", "level", "msg"},
		TimeField:  "time",
		TimeFormat: time.RFC3339,
		LevelField: "level",
		Levels:     []string{"debug", "info", "warn", "error", "fatal"},
	}
}

// loadSchema reads a JSON schema file. Fields left unset (empty string or
// nil slice) fall back to the default schema's value, so a config file
// only needs to mention what it's overriding - e.g. a file with just
// {"levels": ["trace", "info", "error"]} keeps the default required
// fields and time format but replaces the allowed levels.
func loadSchema(path string) (Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Schema{}, fmt.Errorf("reading schema %s: %w", path, err)
	}

	schema := defaultSchema()
	if err := json.Unmarshal(data, &schema); err != nil {
		return Schema{}, fmt.Errorf("parsing schema %s: %w", path, err)
	}
	return schema, nil
}
