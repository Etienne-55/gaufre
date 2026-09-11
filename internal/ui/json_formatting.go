package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderResponse() string {
	resp := m.Response
	if resp.Error != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", resp.Error))
	}
	statusColor := SuccessStyle
	if resp.StatusCode >= 400 {
		statusColor = ErrorStyle
	}
	header := fmt.Sprintf("Status: %s | Time: %v",
		statusColor.Render(fmt.Sprintf("%d", resp.StatusCode)),
		resp.ResponseTime.Round(time.Millisecond))

	formattedBody := m.formatResponseBody(resp.Body)

	return fmt.Sprintf("%s\n\n%s", header, formattedBody)
}

func (m Model) formatResponseBody(body string) string {
	trimmed := strings.TrimSpace(body)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(body), "", "  "); err == nil {
			return highlightJSON(prettyJSON.String())
		}
	}
	return body
}

// stringLiteralRegex matches a single JSON string literal, quotes included.
var stringLiteralRegex = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)

var (
	numberRegex = regexp.MustCompile(`-?\d+\.?\d*`)
	trueRegex   = regexp.MustCompile(`\btrue\b`)
	falseRegex  = regexp.MustCompile(`\bfalse\b`)
	nullRegex   = regexp.MustCompile(`\bnull\b`)
)

// highlightJSON colors pretty-printed JSON for terminal display. Unlike a
// naive regex pass over the whole string, it first locates every string
// literal (keys and values) and only looks for numbers/true/false/null
// outside of those spans - so a string value like "is_nullable" or
// "note: 12 items" never gets partially recolored as if it were a literal
// null or a number.
func highlightJSON(jsonStr string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))    // Blue for keys
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")) // Green for strings
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9e64")) // Orange for numbers
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))   // Purple for booleans
	nullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#545c7e"))   // Gray for null
	punctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5"))  // Light for punctuation

	highlightScalars := func(segment string) string {
		segment = numberRegex.ReplaceAllStringFunc(segment, func(m string) string {
			return numberStyle.Render(m)
		})
		segment = trueRegex.ReplaceAllString(segment, boolStyle.Render("true"))
		segment = falseRegex.ReplaceAllString(segment, boolStyle.Render("false"))
		segment = nullRegex.ReplaceAllString(segment, nullStyle.Render("null"))
		return segment
	}

	var result strings.Builder
	lastEnd := 0

	for _, span := range stringLiteralRegex.FindAllStringIndex(jsonStr, -1) {
		start, end := span[0], span[1]

		result.WriteString(highlightScalars(jsonStr[lastEnd:start]))

		literal := jsonStr[start:end]
		rest := jsonStr[end:]
		trimmedRest := strings.TrimLeft(rest, " \t")

		if strings.HasPrefix(trimmedRest, ":") {
			// This literal is an object key: style it, plus the whitespace
			// and colon that follow, together.
			wsLen := len(rest) - len(trimmedRest)
			result.WriteString(keyStyle.Render(literal + rest[:wsLen]))
			result.WriteString(punctStyle.Render(":"))
			end += wsLen + 1
		} else {
			result.WriteString(stringStyle.Render(literal))
		}

		lastEnd = end
	}

	result.WriteString(highlightScalars(jsonStr[lastEnd:]))

	return result.String()
}
