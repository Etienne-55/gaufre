package ui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// tokenPickerRowsPerItem is how many lines each candidate takes in the
// picker: its source, what it decodes to, a preview of the token itself,
// and the blank line before the next one.
const tokenPickerRowsPerItem = 4

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

// renderTokenCandidates lists the candidates the way the history lists
// requests - same styles, same selection bar - each as its source, what it
// decodes to and the token itself, scrolled so the selected one is always
// on screen when there are more than the terminal can show.
func (m Model) renderTokenCandidates(width int) []string {
	first, last := visibleWindow(len(m.ResponseTokens), m.TokenCursor, m.tokenPickerCapacity())
	now := time.Now()
	current := m.AuthInput.Value()

	styles := list.NewDefaultDelegate().Styles
	textWidth := width - styles.NormalTitle.GetPaddingLeft()

	var rows []string
	for i := first; i < last; i++ {
		c := m.ResponseTokens[i]

		titleStyle, descStyle := styles.NormalTitle, styles.NormalDesc
		if i == m.TokenCursor {
			titleStyle, descStyle = styles.SelectedTitle, styles.SelectedDesc
		}

		title := c.Source
		if c.Token == current {
			title += " · in use"
		}

		// An expired or expiring token keeps its warning color whether or
		// not it is selected.
		summary, summaryStyle := describeToken(c.Token, now)
		summaryLineStyle := descStyle
		if fg := summaryStyle.GetForeground(); fg == ColorBad || fg == ColorWarn {
			summaryLineStyle = descStyle.Foreground(fg)
		}

		if i > first {
			rows = append(rows, "")
		}
		rows = append(rows,
			titleStyle.Render(truncateString(title, textWidth)),
			summaryLineStyle.Render(truncateString(summary, textWidth)),
			descStyle.Render(truncateMiddle(c.Token, textWidth)),
		)
	}

	return rows
}

// truncateMiddle shortens s to max cells by cutting out its middle. Tokens
// from one response often share a prefix - every JWT starts with much the
// same header - so keeping the end as well tells them apart.
func truncateMiddle(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max || max < 3 {
		return truncateString(s, max)
	}
	head := max / 2
	tail := max - head - 1 // one cell for the ellipsis
	return string(runes[:head]) + "…" + string(runes[len(runes)-tail:])
}

func (m Model) renderTokenPickerStatus(width int) string {
	if m.Status != "" {
		return EditorOKStyle.Render(truncateString(m.Status, width))
	}
	return EditorMetaStyle.Render(truncateString(
		fmt.Sprintf("%d of %d", m.TokenCursor+1, len(m.ResponseTokens)), width))
}

// tokenPickerCapacity is how many candidates fit in the picker frame on the
// current terminal. The last one has no blank line after it, hence the +1.
// A negative result means the height is unknown.
func (m Model) tokenPickerCapacity() int {
	if m.Height <= 0 {
		return -1
	}
	return maxInt(1, (m.Height-editorChromeRows+1)/tokenPickerRowsPerItem)
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
