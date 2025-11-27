package ui

import (
	"fmt"
	"time"
	"bytes"
	"regexp"
	"strings"
	"encoding/json"
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
			return m.highlightJSON(prettyJSON.String())
		}
	}
	return body
}

func (m Model) highlightJSON(jsonStr string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))     // Blue for keys
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))  // Green for strings
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9e64"))  // Orange for numbers
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))    // Purple for booleans
	nullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#545c7e"))    // Gray for null
	punctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5"))   // Light for punctuation
	
	result := jsonStr
	
	keyRegex := regexp.MustCompile(`"([^"]+)"\s*:`)
	result = keyRegex.ReplaceAllStringFunc(result, func(match string) string {
		return keyStyle.Render(match[:len(match)-1]) + punctStyle.Render(":")
	})
	
	stringRegex := regexp.MustCompile(`:\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	result = stringRegex.ReplaceAllStringFunc(result, func(match string) string {
		colonIdx := strings.Index(match, ":")
		firstQuoteIdx := strings.Index(match[colonIdx:], `"`) + colonIdx
		lastQuoteIdx := strings.LastIndex(match, `"`)
		
		prefix := match[:firstQuoteIdx]
		stringContent := match[firstQuoteIdx : lastQuoteIdx+1]
		
		return prefix + stringStyle.Render(stringContent)
	})
	
	numberRegex := regexp.MustCompile(`:\s*(-?\d+\.?\d*)`)
	result = numberRegex.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, ":", 2)
		return parts[0] + ":" + numberStyle.Render(strings.TrimSpace(parts[1]))
	})
	
	result = regexp.MustCompile(`\btrue\b`).ReplaceAllString(result, boolStyle.Render("true"))
	result = regexp.MustCompile(`\bfalse\b`).ReplaceAllString(result, boolStyle.Render("false"))
	result = regexp.MustCompile(`\bnull\b`).ReplaceAllString(result, nullStyle.Render("null"))
	
	return result
}

