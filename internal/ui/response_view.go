package ui

import (
	"fmt"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Rows the response frame spends outside the viewport: two borders, the
	// vertical padding, the pinned status header and the blank line under
	// it, the action status line and the help line.
	responseChromeRows = 8

	responseMinViewportHeight = 3
)

// responseBoxWidth is the width of the response frame, border excluded.
// Prefers a margin around the frame and a readable maximum, but never
// outgrows a narrow terminal (it used to insist on at least 80 columns).
func (m Model) responseBoxWidth() int {
	preferred := clampInt(m.Width-10, 80, 150)
	return maxInt(minInt(preferred, m.Width-2), 20)
}

// responseViewportSize is the area left for the scrolling body once the
// frame's padding, header and footer are accounted for.
func (m Model) responseViewportSize() (int, int) {
	width := m.responseBoxWidth() - 4 // horizontal padding
	height := maxInt(m.Height-responseChromeRows-1, responseMinViewportHeight)
	return width, height
}

// renderResponseContent is what the viewport scrolls: optionally the
// headers, then the body. It is wrapped to the viewport width here, rather
// than left for the frame to re-wrap, so the viewport's line count matches
// what is drawn and a long line (a JWT, minified JSON) can't push the frame
// past the bottom of the terminal.
func (m Model) renderResponseContent(width int) string {
	resp := m.Response
	if resp == nil {
		return ""
	}

	var content string
	if resp.Error != nil {
		content = ErrorStyle.Render(fmt.Sprintf("Error: %v", resp.Error))
	} else {
		content = m.formatResponseBody(resp.Body)
		if strings.TrimSpace(resp.Body) == "" {
			content = EditorMetaStyle.Render("(empty body)")
		}
		if m.ShowHeaders {
			content = renderResponseHeaders(resp.Headers) + "\n\n" + content
		}
	}

	return lipgloss.NewStyle().Width(width).Render(content)
}

func renderResponseHeaders(headers map[string][]string) string {
	if len(headers) == 0 {
		return EditorMetaStyle.Render("(no headers)")
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
	var lines []string
	for _, name := range sortedKeys(headers) {
		for _, value := range headers[name] {
			lines = append(lines, keyStyle.Render(name+":")+" "+EditorValueStyle.Render(value))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderResponseScreen() string {
	width := m.Viewport.Width

	help := m.Help
	help.Width = width

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.renderResponseHeader(width),
		"",
		m.Viewport.View(),
		m.renderResponseStatus(width),
		help.View(m.responseHelp()),
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Width(m.responseBoxWidth()).
		Padding(1, 2)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

// renderResponseHeader is pinned above the body, so the status stays in
// view while scrolling a long response: "200 OK · 123ms · 2.4 KB ·
// application/json" on the left, the scroll position on the right.
func (m Model) renderResponseHeader(width int) string {
	resp := m.Response
	elapsed := resp.ResponseTime.Round(time.Millisecond).String()

	var left string
	if resp.Error != nil {
		left = ErrorStyle.Render("Request failed") + EditorMetaStyle.Render(" · "+elapsed)
	} else {
		meta := []string{elapsed, formatBytes(len(resp.Body))}
		if contentType := mediaType(resp.Headers); contentType != "" {
			meta = append(meta, contentType)
		}
		left = statusCodeStyle(resp.StatusCode).Render(statusLine(resp.StatusCode)) +
			EditorMetaStyle.Render(" · "+strings.Join(meta, " · "))
	}

	right := ""
	if total := m.Viewport.TotalLineCount(); total > m.Viewport.Height {
		first := m.Viewport.YOffset + 1
		last := minInt(m.Viewport.YOffset+m.Viewport.Height, total)
		right = EditorMetaStyle.Render(fmt.Sprintf("%d–%d of %d", first, last, total))
	}

	left = truncateStyled(left, width-lipgloss.Width(right)-1)
	gap := maxInt(width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

// renderResponseStatus fills the line above the help with the result of
// the last action (token set, nothing found, ...), so the frame keeps the
// same height either way.
func (m Model) renderResponseStatus(width int) string {
	if m.Status == "" {
		return ""
	}
	return EditorOKStyle.Render(truncateString(m.Status, width))
}

// responseHelp only lists the shortcuts that do something for this
// response - the token one in particular doubles as a "there's a token
// here" hint.
func (m Model) responseHelp() helpBindings {
	bindings := helpBindings{}
	if m.Viewport.TotalLineCount() > m.Viewport.Height {
		bindings = append(bindings, ResponseKeys.Scroll, ResponseKeys.Top)
	}

	if n := len(m.ResponseTokens); n > 0 {
		useToken := ResponseKeys.UseToken
		if n > 1 {
			useToken.SetHelp("t", fmt.Sprintf("pick token (%d)", n))
		}
		bindings = append(bindings, useToken)
	}

	if m.Response.Error == nil {
		headers := ResponseKeys.Headers
		if m.ShowHeaders {
			headers.SetHelp("h", "hide headers")
		}
		bindings = append(bindings, headers)
	}

	return append(bindings, ResponseKeys.Back, ResponseKeys.Quit)
}

// statusLine renders a status code with its reason phrase: "404 Not Found".
func statusLine(code int) string {
	if text := nethttp.StatusText(code); text != "" {
		return fmt.Sprintf("%d %s", code, text)
	}
	return fmt.Sprintf("%d", code)
}

func statusCodeStyle(code int) lipgloss.Style {
	switch {
	case code >= 400:
		return ErrorStyle
	case code >= 300:
		return EditorWarnStyle.Bold(true)
	default:
		return SuccessStyle
	}
}

// mediaType is the response's Content-Type without its parameters
// ("application/json", not "application/json; charset=utf-8").
func mediaType(headers map[string][]string) string {
	values := nethttp.Header(headers).Values("Content-Type")
	if len(values) == 0 {
		return ""
	}
	mediaType, _, _ := strings.Cut(values[0], ";")
	return strings.TrimSpace(mediaType)
}

func formatBytes(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
