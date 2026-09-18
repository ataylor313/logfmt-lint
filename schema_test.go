package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSchemaOverridesOnlyGivenFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(path, []byte(`{"levels": ["trace", "info", "error"]}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := loadSchema(path)
	if err != nil {
		t.Fatalf("loadSchema: %v", err)
	}

	want := defaultSchema()
	want.Levels = []string{"trace", "info", "error"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("loadSchema(%q) = %#v, want %#v", path, got, want)
	}
}

func TestLoadSchemaMissingFile(t *testing.T) {
	if _, err := loadSchema(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("loadSchema on a missing file returned nil error, want an error")
	}
}

func TestLoadSchemaInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := loadSchema(path); err == nil {
		t.Error("loadSchema on invalid JSON returned nil error, want an error")
	}
}

func TestValidateWithCustomSchema(t *testing.T) {
	schema := Schema{
		Required:   []string{"ts", "severity"},
		TimeField:  "ts",
		TimeFormat: "2006-01-02",
		LevelField: "severity",
		Levels:     []string{"low", "high"},
	}

	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid", "ts=2024-01-15 severity=high", false},
		{"bad time format", "ts=2024-01-15T10:00:00Z severity=high", true},
		{"bad level", "ts=2024-01-15 severity=medium", true},
		{"missing required field", "ts=2024-01-15", true},
		{"default schema fields ignored", "time=2024-01-15T10:00:00Z level=info ts=2024-01-15 severity=high", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parseLine(tc.raw)
			if err != nil {
				t.Fatalf("parseLine(%q): %v", tc.raw, err)
			}
			errs := validate(entry, schema)
			if tc.wantErr && len(errs) == 0 {
				t.Errorf("validate(%q) = no errors, want at least one", tc.raw)
			}
			if !tc.wantErr && len(errs) != 0 {
				t.Errorf("validate(%q) = %v, want no errors", tc.raw, errs)
			}
		})
	}
}

func TestValidateDefaultSchemaUnchanged(t *testing.T) {
	entry, err := parseLine(`time=2024-01-15T10:30:00Z level=info msg="request completed"`)
	if err != nil {
		t.Fatalf("parseLine: %v", err)
	}
	if errs := validate(entry, defaultSchema()); len(errs) != 0 {
		t.Errorf("validate() with default schema = %v, want no errors", errs)
	}
}
