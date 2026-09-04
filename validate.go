package main

import (
	"fmt"
	"time"
)

// The schema is fixed for now: every line is expected to carry a
// timestamp, a level from a known set, and a message. A configurable
// schema is on the roadmap; this is enough to catch the common case of a
// log line that silently dropped a field.
var allowedLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"fatal": true,
}

func validate(e Entry) []string {
	var errs []string

	if ts, ok := e.Get("time"); !ok {
		errs = append(errs, "missing required field: time")
	} else if _, err := time.Parse(time.RFC3339, ts); err != nil {
		errs = append(errs, fmt.Sprintf("field time is not RFC3339: %q", ts))
	}

	if level, ok := e.Get("level"); !ok {
		errs = append(errs, "missing required field: level")
	} else if !allowedLevels[level] {
		errs = append(errs, fmt.Sprintf("field level has unrecognized value: %q", level))
	}

	if _, ok := e.Get("msg"); !ok {
		errs = append(errs, "missing required field: msg")
	}

	return errs
}
