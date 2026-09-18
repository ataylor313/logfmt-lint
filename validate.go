package main

import (
	"fmt"
	"time"
)

func validate(e Entry, schema Schema) []string {
	var errs []string

	levels := make(map[string]bool, len(schema.Levels))
	for _, l := range schema.Levels {
		levels[l] = true
	}

	for _, field := range schema.Required {
		val, ok := e.Get(field)
		if !ok {
			errs = append(errs, fmt.Sprintf("missing required field: %s", field))
			continue
		}

		switch {
		case field == schema.TimeField && schema.TimeFormat != "":
			if _, err := time.Parse(schema.TimeFormat, val); err != nil {
				errs = append(errs, fmt.Sprintf("field %s does not match format %q: %q", field, schema.TimeFormat, val))
			}
		case field == schema.LevelField && len(levels) > 0:
			if !levels[val] {
				errs = append(errs, fmt.Sprintf("field %s has unrecognized value: %q", field, val))
			}
		}
	}

	return errs
}
