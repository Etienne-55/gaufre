package ui

import (
	"strings"
	"testing"
)

func TestInspectBody(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantEmpty bool
		wantValid bool
		wantLines int
	}{
		{name: "empty", body: "", wantEmpty: true, wantLines: 1},
		{name: "whitespace only", body: "  \n ", wantEmpty: true, wantLines: 2},
		{name: "valid object", body: "{\n  \"a\": 1\n}", wantValid: true, wantLines: 3},
		{name: "valid array", body: `[1, 2, 3]`, wantValid: true, wantLines: 1},
		{name: "invalid", body: `{"a": }`, wantLines: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inspectBody(tt.body)
			if got.Empty != tt.wantEmpty {
				t.Errorf("Empty = %v, want %v", got.Empty, tt.wantEmpty)
			}
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v (problem: %s)", got.Valid, tt.wantValid, got.Problem)
			}
			if got.Lines != tt.wantLines {
				t.Errorf("Lines = %d, want %d", got.Lines, tt.wantLines)
			}
			if got.Bytes != len(tt.body) {
				t.Errorf("Bytes = %d, want %d", got.Bytes, len(tt.body))
			}
		})
	}
}

func TestInspectBody_ReportsLineAndColumn(t *testing.T) {
	body := "{\n  \"name\": \"test\",\n  \"broken\": \n}"

	got := inspectBody(body)
	if got.Valid {
		t.Fatal("expected invalid JSON")
	}
	if !strings.Contains(got.Problem, "line 4") {
		t.Errorf("Problem = %q, want it to mention line 4", got.Problem)
	}
	if !strings.Contains(got.Problem, "col ") {
		t.Errorf("Problem = %q, want it to mention a column", got.Problem)
	}
}

func TestFormatJSONBody(t *testing.T) {
	got, err := formatJSONBody(`{"a":1,"b":[2,3]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "{\n  \"a\": 1,\n  \"b\": [\n    2,\n    3\n  ]\n}"
	if got != want {
		t.Errorf("formatJSONBody() =\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatJSONBody_LeavesInvalidBodyAlone(t *testing.T) {
	body := `{"a": }`

	got, err := formatJSONBody(body)
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
	if got != body {
		t.Errorf("body was modified despite the error: %q", got)
	}
}

func TestFormatJSONBody_Empty(t *testing.T) {
	got, err := formatJSONBody("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "   " {
		t.Errorf("formatJSONBody(%q) = %q", "   ", got)
	}
}

func TestLineAndColumn(t *testing.T) {
	body := "ab\ncde\nf"

	tests := []struct {
		offset   int64
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{2, 1, 3},
		{3, 2, 1},
		{6, 2, 4},
		{7, 3, 1},
		{100, 3, 2}, // past the end is clamped
	}

	for _, tt := range tests {
		line, col := lineAndColumn(body, tt.offset)
		if line != tt.wantLine || col != tt.wantCol {
			t.Errorf("lineAndColumn(offset=%d) = (%d, %d), want (%d, %d)", tt.offset, line, col, tt.wantLine, tt.wantCol)
		}
	}
}
