package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Shared palette for the editor screens.
var (
	ColorAccent     = lipgloss.Color("#4ECDC4")
	ColorBorder     = lipgloss.Color("#874BFD")
	ColorText       = lipgloss.Color("#c0caf5")
	ColorMuted      = lipgloss.Color("#545c7e")
	ColorBackground = lipgloss.Color("#24283b")
	ColorOK         = lipgloss.Color("#9ece6a")
	ColorWarn       = lipgloss.Color("#e0af68")
	ColorBad        = lipgloss.Color("#f7768e")
)

// Bounds for the editor frames. Every child of an editor frame is rendered
// at exactly the frame's content width: lipgloss re-wraps any child wider
// than its container, which is what used to shred the token editor's box
// drawing when a long token soft-wrapped.
const (
	// authPrompt doubles as the token field's prompt and as a reminder of
	// the header that will actually be sent.
	authPrompt = "Bearer "

	editorMinContentWidth = 40
	editorMaxContentWidth = 96
	editorMinBodyHeight   = 4
	editorMaxBodyHeight   = 32

	// Space the frame gives up to its own border and padding, plus the
	// margin kept between the frame and the edge of the terminal when
	// there is room for it.
	editorFrameChrome  = 4 // left/right border + padding
	editorScreenMargin = 6

	// Rows every editor frame spends on chrome: two borders, the title, the
	// blank lines around the field, the status line and the help line.
	editorChromeRows = 7

	// The token editor additionally has the input line and the blank line
	// under it above its detail panel.
	authChromeRows = editorChromeRows + 2
)

var (
	EditorTitleStyle = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	EditorMetaStyle  = lipgloss.NewStyle().Foreground(ColorMuted)
	EditorLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	EditorValueStyle = lipgloss.NewStyle().Foreground(ColorText)
	EditorOKStyle    = lipgloss.NewStyle().Foreground(ColorOK)
	EditorWarnStyle  = lipgloss.NewStyle().Foreground(ColorWarn)
	EditorBadStyle   = lipgloss.NewStyle().Foreground(ColorBad)
)

// editorContentWidth is the usable width inside an editor frame for the
// current terminal size. The preferred width keeps a margin around the
// frame, but a narrow terminal gives that margin up rather than letting the
// frame overflow the screen.
func (m Model) editorContentWidth() int {
	if m.Width <= 0 {
		return editorMinContentWidth // no WindowSizeMsg yet
	}

	preferred := clampInt(m.Width-editorScreenMargin-editorFrameChrome, editorMinContentWidth, editorMaxContentWidth)
	fits := m.Width - editorFrameChrome

	return clampInt(minInt(preferred, fits), 1, editorMaxContentWidth)
}

// editorBodyHeight is how many rows the body textarea gets, leaving room
// for the frame's title, status line, help line and the blank rows between
// them.
func (m Model) editorBodyHeight() int {
	if m.Height <= 0 {
		return editorMinBodyHeight
	}

	preferred := clampInt(m.Height-editorScreenMargin-editorChromeRows, editorMinBodyHeight, editorMaxBodyHeight)
	fits := m.Height - editorChromeRows

	return clampInt(minInt(preferred, fits), 1, editorMaxBodyHeight)
}

// authDetailRows is how many rows the token editor can spend on its decoded
// JWT panel before the frame would outgrow the terminal. A negative result
// means the height is unknown and the panel is unconstrained.
func (m Model) authDetailRows() int {
	if m.Height <= 0 {
		return -1
	}
	return maxInt(0, m.Height-authChromeRows)
}

// renderEditorFrame draws the single bordered box shared by both editors and
// centers it in the terminal. width is the width available to the rows:
// lipgloss counts padding inside Width, so the frame is widened by its own
// horizontal padding to leave exactly that much room.
func (m Model) renderEditorFrame(width int, rows ...string) string {
	const framePadding = 1

	frame := lipgloss.NewStyle().
		Border(m.screenBorder()).
		BorderForeground(ColorBorder).
		Padding(0, framePadding).
		Width(width + 2*framePadding)

	return lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		frame.Render(lipgloss.JoinVertical(lipgloss.Left, rows...)),
	)
}

// renderTitleBar lays out a left-aligned title and right-aligned context
// (method + URL) on one line of exactly width cells, truncating the context
// rather than letting the line overflow.
func renderTitleBar(width int, title, context string) string {
	left := EditorTitleStyle.Render(title)
	leftWidth := lipgloss.Width(left)

	available := width - leftWidth - 1
	if available < 0 {
		available = 0
	}

	right := EditorMetaStyle.Render(truncateString(context, available))
	gap := width - leftWidth - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

// renderFieldRow renders an aligned "label  value" row, used for the decoded
// JWT claims.
func renderFieldRow(width, labelWidth int, label, value string, valueStyle lipgloss.Style) string {
	paddedLabel := label
	if len(paddedLabel) < labelWidth {
		paddedLabel += strings.Repeat(" ", labelWidth-len(paddedLabel))
	}

	rendered := EditorLabelStyle.Render(paddedLabel) + " "
	return rendered + valueStyle.Render(truncateString(value, width-lipgloss.Width(rendered)))
}

// truncateString shortens s to at most max cells, marking the cut with an
// ellipsis.
func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// trimRows shortens rows to at most budget entries, replacing the last kept
// entry with an ellipsis so it's clear something was cut. A negative budget
// leaves rows untouched.
func trimRows(rows []string, budget int) []string {
	if budget < 0 || len(rows) <= budget {
		return rows
	}
	if budget == 0 {
		return nil
	}

	trimmed := make([]string, 0, budget)
	trimmed = append(trimmed, rows[:budget-1]...)
	return append(trimmed, EditorMetaStyle.Render("…"))
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
