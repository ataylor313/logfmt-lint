package main

import (
	"reflect"
	"testing"
)

func TestParseLineOK(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []Field
	}{
		{
			name: "empty line",
			raw:  "",
			want: nil,
		},
		{
			name: "single bare pair",
			raw:  "level=info",
			want: []Field{{Key: "level", Value: "info"}},
		},
		{
			name: "multiple pairs with extra spacing",
			raw:  "  level=info   msg=hi  ",
			want: []Field{{Key: "level", Value: "info"}, {Key: "msg", Value: "hi"}},
		},
		{
			name: "quoted value with space",
			raw:  `msg="cache miss"`,
			want: []Field{{Key: "msg", Value: "cache miss"}},
		},
		{
			name: "quoted value with escaped quote and backslash",
			raw:  `msg="say \"hi\" then \\ back off"`,
			want: []Field{{Key: "msg", Value: `say "hi" then \ back off`}},
		},
		{
			name: "empty bare value",
			raw:  "key=",
			want: []Field{{Key: "key", Value: ""}},
		},
		{
			name: "empty quoted value",
			raw:  `key=""`,
			want: []Field{{Key: "key", Value: ""}},
		},
		{
			name: "unescaped equals inside bare value",
			raw:  "a=b=c",
			want: []Field{{Key: "a", Value: "b=c"}},
		},
		{
			name: "mixed bare and quoted",
			raw:  `time=2024-01-15T10:30:00Z msg="request completed" status=200`,
			want: []Field{
				{Key: "time", Value: "2024-01-15T10:30:00Z"},
				{Key: "msg", Value: "request completed"},
				{Key: "status", Value: "200"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parseLine(tc.raw)
			if err != nil {
				t.Fatalf("parseLine(%q) returned error: %v", tc.raw, err)
			}
			if !reflect.DeepEqual(entry.Fields, tc.want) {
				t.Errorf("parseLine(%q).Fields = %#v, want %#v", tc.raw, entry.Fields, tc.want)
			}
			if entry.Raw != tc.raw {
				t.Errorf("parseLine(%q).Raw = %q, want %q", tc.raw, entry.Raw, tc.raw)
			}
		})
	}
}

func TestParseLineErrors(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"missing equals", "justakey"},
		{"missing equals after valid pair", "a=b justakey"},
		{"empty key", "=value"},
		{"unterminated quote", `msg="cache miss`},
		{"dangling backslash inside quotes", `msg="trailing\`},
		{"invalid escape sequence", `msg="bad \q escape"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseLine(tc.raw); err == nil {
				t.Errorf("parseLine(%q) returned nil error, want an error", tc.raw)
			}
		})
	}
}

func TestEntryGet(t *testing.T) {
	entry, err := parseLine("level=info level=debug msg=hi")
	if err != nil {
		t.Fatalf("parseLine returned error: %v", err)
	}

	if v, ok := entry.Get("level"); !ok || v != "info" {
		t.Errorf("Get(%q) = (%q, %v), want (%q, true) - first match should win on duplicate keys", "level", v, ok, "info")
	}
	if v, ok := entry.Get("msg"); !ok || v != "hi" {
		t.Errorf("Get(%q) = (%q, %v), want (%q, true)", "msg", v, ok, "hi")
	}
	if _, ok := entry.Get("missing"); ok {
		t.Errorf("Get(%q) reported found, want not found", "missing")
	}
}
