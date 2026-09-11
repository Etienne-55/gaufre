package ui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tokenPickerRowsPerItem is how many lines each candidate takes in the
// picker: its source and summary, then a preview of the token itself.
const tokenPickerRowsPerItem = 2

// openTokenPicker is what "t" does in the response viewer. With a single
// token in the response - the usual login response - it goes straight into
// the auth editor; with several (access + refresh + id token, say) it opens
// the picker so you can choose.
func (m Model) openTokenPicker() (tea.Model, tea.Cmd) {
	switch len(m.ResponseTokens) {
	case 0:
		m, cmd := m.setStatus("no token found in this response")
		return m, cmd
	case 1:
		return m.useResponseToken(m.ResponseTokens[0])
	}

	m.TokenCursor = 0
	m.Status = ""
	m.Screen = ScreenTokenPicker
	return m, nil
}

// useResponseToken makes a token from the response the bearer token for
// every following request, and returns to the response viewer.
func (m Model) useResponseToken(c tokenCandidate) (tea.Model, tea.Cmd) {
	m.AuthInput.SetValue(c.Token)
	// Show the start of the token: its head is what identifies it.
	m.AuthInput.CursorStart()
	m.Screen = ScreenResponse

	status := "auth token set from " + c.Source
	if _, isJWT := decodeJWT(c.Token); isJWT {
		status = "auth token set from " + c.Source + " (JWT)"
	}
	m, cmd := m.setStatus(status)
	return m, cmd
}

func (m Model) updateTokenPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || len(m.ResponseTokens) == 0 {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, TokenPickerKeys.Back):
		m.Status = ""
		m.Screen = ScreenResponse
		return m, nil

	case key.Matches(keyMsg, TokenPickerKeys.Up):
		if m.TokenCursor > 0 {
			m.TokenCursor--
		}
		return m, nil

	case key.Matches(keyMsg, TokenPickerKeys.Down):
		if m.TokenCursor < len(m.ResponseTokens)-1 {
			m.TokenCursor++
		}
		return m, nil

	case key.Matches(keyMsg, TokenPickerKeys.Use):
		return m.useResponseToken(m.ResponseTokens[m.TokenCursor])

	case key.Matches(keyMsg, TokenPickerKeys.Copy):
		c := m.ResponseTokens[m.TokenCursor]
		if err := clipboard.WriteAll(c.Token); err != nil {
			m, cmd := m.setStatus("copy failed: " + err.Error())
			return m, cmd
		}
		m, cmd := m.setStatus(c.Source + " copied to clipboard")
		return m, cmd
	}

	return m, nil
}

func (m Model) renderTokenPickerScreen() string {
	width := m.editorContentWidth()

	help := m.Help
	help.Width = width

	title := fmt.Sprintf("Tokens in response (%d)", len(m.ResponseTokens))
	rows := []string{renderTitleBar(width, title, m.requestContext()), ""}
	rows = append(rows, m.renderTokenCandidates(width)...)
	rows = append(rows, "", m.renderTokenPickerStatus(width), help.View(TokenPickerKeys))

	return m.renderEditorFrame(width, rows...)
}

// renderTokenCandidates lists the candidates, scrolled so the selected one
// is always on screen when there are more than the terminal can show.
func (m Model) renderTokenCandidates(width int) []string {
	first, last := visibleWindow(len(m.ResponseTokens), m.TokenCursor, m.tokenPickerCapacity())
	now := time.Now()
	current := m.AuthInput.Value()

	var rows []string
	for i := first; i < last; i++ {
		c := m.ResponseTokens[i]
		selected := i == m.TokenCursor

		marker := "  "
		sourceStyle := EditorValueStyle
		if selected {
			marker = EditorTitleStyle.Render("› ")
			sourceStyle = EditorTitleStyle
		}

		summary, summaryStyle := describeToken(c.Token, now)
		if c.Token == current {
			summary += " · in use"
		}

		// Source on the left, summary on the right, like the editors'
		// title bar; the source gives way first on a narrow terminal.
		source := sourceStyle.Render(truncateString(c.Source, maxInt(width-2-lipgloss.Width(summary)-2, 8)))
		right := summaryStyle.Render(truncateString(summary, maxInt(width-2-lipgloss.Width(source)-1, 0)))
		gap := maxInt(width-2-lipgloss.Width(source)-lipgloss.Width(right), 1)

		rows = append(rows,
			marker+source+fmt.Sprintf("%*s", gap, "")+right,
			"  "+EditorMetaStyle.Render(truncateString(c.Token, width-2)),
		)
	}

	return rows
}

func (m Model) renderTokenPickerStatus(width int) string {
	if m.Status != "" {
		return EditorOKStyle.Render(truncateString(m.Status, width))
	}
	return EditorMetaStyle.Render(truncateString(
		fmt.Sprintf("%d of %d", m.TokenCursor+1, len(m.ResponseTokens)), width))
}

// tokenPickerCapacity is how many candidates fit in the picker frame on the
// current terminal. A negative result means the height is unknown.
func (m Model) tokenPickerCapacity() int {
	if m.Height <= 0 {
		return -1
	}
	return maxInt(1, (m.Height-editorChromeRows)/tokenPickerRowsPerItem)
}

// visibleWindow returns the [first, last) range of a list of n items to
// show when only capacity fit, keeping cursor inside it. A negative
// capacity shows everything.
func visibleWindow(n, cursor, capacity int) (int, int) {
	if capacity < 0 || n <= capacity {
		return 0, n
	}
	first := clampInt(cursor-capacity/2, 0, n-capacity)
	return first, first + capacity
}
