package main

import (
	"encoding/json"
	"testing"
)

func TestRedactAdjacentIPv4(t *testing.T) {
	r := newRedactor()
	got := r.redactString("45.67.89.10,56.78.90.11 192.168.8.1")
	want := "203.0.113.1,203.0.113.2 192.168.8.1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedactIPv4InsideLongerNumber(t *testing.T) {
	r := newRedactor()
	for _, s := range []string{"1.2.3.4.5", "1234.5.6.7", "4.7.0"} {
		if got := r.redactString(s); got != s {
			t.Errorf("redactString(%q) = %q", s, got)
		}
	}
}

func TestRedactPinCode(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"profiles": []any{map[string]any{"pin_code": "1234", "apn": "internet"}}}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"profiles":[{"apn":"internet","pin_code":"redacted-pin_code-1"}]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestScrubKeepsJSONValid(t *testing.T) {
	r := newRedactor()
	r.addSecret("true", "password")
	v := map[string]any{"enabled": true, "note": "is true", "count": float64(3)}
	out := mustJSON(r.scrub(r.redact(v)))

	var decoded map[string]any
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["enabled"] != true {
		t.Errorf("enabled = %v, want true", decoded["enabled"])
	}
	if decoded["note"] != "is redacted-password-1" {
		t.Errorf("note = %v", decoded["note"])
	}
}
