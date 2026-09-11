package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

var ansiCodeRegex = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiCodeRegex.ReplaceAllString(s, "")
}

func TestHighlightJSON_PreservesContent(t *testing.T) {
	input := `{
  "name": "test",
  "count": 12,
  "active": true,
  "missing": null
}`
	got := highlightJSON(input)
	if stripANSI(got) != input {
		t.Fatalf("highlighting changed the underlying text:\n got: %q\nwant: %q", stripANSI(got), input)
	}
}

func TestHighlightJSON_DoesNotMiscolorTokensInsideStrings(t *testing.T) {
	// "true", "null" and digits appearing inside a *string value* must not
	// be wrapped as if they were literal JSON tokens: the whole string
	// value should come out as one contiguous span in the string style.
	input := `{
  "note": "is_nullable is true, item 12 of 100"
}`
	got := highlightJSON(input)

	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	wantValue := stringStyle.Render(`"is_nullable is true, item 12 of 100"`)

	if !strings.Contains(got, wantValue) {
		t.Errorf("string value was not rendered as a single contiguous span (true/number/null likely miscolored inside it)\n got: %q\nwant substring: %q", got, wantValue)
	}
}

func TestFormatResponseBody_NonJSONPassthrough(t *testing.T) {
	m := Model{}
	body := "plain text response, not json"
	if got := m.formatResponseBody(body); got != body {
		t.Errorf("formatResponseBody(%q) = %q, want unchanged", body, got)
	}
}

func TestFormatResponseBody_PrettyPrintsJSON(t *testing.T) {
	m := Model{}
	got := m.formatResponseBody(`{"a":1}`)
	if !strings.Contains(stripANSI(got), "\"a\": 1") {
		t.Errorf("formatResponseBody did not indent JSON: %q", stripANSI(got))
	}
}
