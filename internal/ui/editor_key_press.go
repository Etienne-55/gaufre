package ui

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// statusExpiredMsg clears a transient editor status message. Seq identifies
// which message the tick belongs to, so an older tick can't clear a newer
// message.
type statusExpiredMsg struct{ seq int }

const statusLifetime = 2 * time.Second

// setStatus shows a transient message in the editor status line and
// schedules its removal.
func (m Model) setStatus(text string) (Model, tea.Cmd) {
	m.StatusSeq++
	m.Status = text
	seq := m.StatusSeq

	return m, tea.Tick(statusLifetime, func(time.Time) tea.Msg {
		return statusExpiredMsg{seq: seq}
	})
}

// updatePayloadEditor drives the JSON body editor. Anything that isn't one
// of the shortcuts below is forwarded to the textarea, which owns cursor
// movement, scrolling, wrapping and its own paste handling.
func (m Model) updatePayloadEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, PayloadKeys.Back):
			m.PayloadArea.Blur()
			m.Status = ""
			m.Screen = ScreenMain
			m.MainFocus = FocusURL
			return m, nil

		case key.Matches(keyMsg, PayloadKeys.Format):
			formatted, err := formatJSONBody(m.PayloadArea.Value())
			if err != nil {
				m, cmd := m.setStatus("can't format: " + err.Error())
				return m, cmd
			}
			m.PayloadArea.SetValue(formatted)
			m, cmd := m.setStatus("formatted")
			return m, cmd

		case key.Matches(keyMsg, PayloadKeys.Indent):
			// The textarea sanitizes literal tabs into spaces anyway, so
			// insert the indent the formatter uses and keep it honest.
			m.PayloadArea.InsertString("  ")
			return m, nil

		case key.Matches(keyMsg, PayloadKeys.Clear):
			m.PayloadArea.Reset()
			m, cmd := m.setStatus("body cleared")
			return m, cmd

		case key.Matches(keyMsg, PayloadKeys.Copy):
			if err := clipboard.WriteAll(m.PayloadArea.Value()); err != nil {
				m, cmd := m.setStatus("copy failed: " + err.Error())
				return m, cmd
			}
			m, cmd := m.setStatus("body copied to clipboard")
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.PayloadArea, cmd = m.PayloadArea.Update(msg)
	return m, cmd
}

// updateAuthEditor drives the bearer token field. Pasting runs the clipboard
// through cleanAuthToken first, so a token copied straight out of an API
// response - JSON wrapper, "Bearer " prefix, line breaks and all - lands as
// a usable value.
func (m Model) updateAuthEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, AuthKeys.Back):
			m.AuthInput.Blur()
			m.Status = ""
			m.Screen = ScreenMain
			m.MainFocus = FocusPayloadMenu
			return m, nil

		case key.Matches(keyMsg, AuthKeys.Clear):
			m.AuthInput.Reset()
			m, cmd := m.setStatus("token cleared")
			return m, cmd

		case key.Matches(keyMsg, AuthKeys.Copy):
			if err := clipboard.WriteAll(m.AuthInput.Value()); err != nil {
				m, cmd := m.setStatus("copy failed: " + err.Error())
				return m, cmd
			}
			m, cmd := m.setStatus("token copied to clipboard")
			return m, cmd

		case key.Matches(keyMsg, AuthKeys.Paste):
			text, err := clipboard.ReadAll()
			if err != nil {
				m, cmd := m.setStatus("paste failed: " + err.Error())
				return m, cmd
			}

			token := cleanAuthToken(text)
			if token == "" {
				m, cmd := m.setStatus("clipboard held no usable token")
				return m, cmd
			}

			m.AuthInput.SetValue(token)
			// Show the start of the token: its head is what identifies it.
			m.AuthInput.CursorStart()

			status := "token pasted"
			if _, isJWT := decodeJWT(token); isJWT {
				status = "JWT pasted and decoded"
			}
			m, cmd := m.setStatus(status)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.AuthInput, cmd = m.AuthInput.Update(msg)
	return m, cmd
}
