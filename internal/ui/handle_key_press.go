package ui

import (
	"gaufre/internal/http"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMainKeyPress handles input for the main screen only (method
// selector, URL bar, Body/Auth menu). Payload/Auth editing, history and the
// response viewer each own their input handling in their own file, since
// each is really a separate screen with its own key semantics.
func (m Model) handleMainKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q":
		return m, tea.Quit

	case "tab":
		m.Screen = ScreenHistory
		return m, nil

	case "ctrl+v":
		if m.MainFocus == FocusURL {
			if text, err := clipboard.ReadAll(); err == nil {
				m.URL = m.URL[:m.Cursor] + text + m.URL[m.Cursor:]
				m.Cursor += len(text)
			}
		}
		return m, nil

	case "enter":
		return m.handleMainEnter()

	case "backspace":
		if m.MainFocus == FocusURL && len(m.URL) > 0 && m.Cursor > 0 {
			m.URL = m.URL[:m.Cursor-1] + m.URL[m.Cursor:]
			m.Cursor--
		}
		return m, nil

	case "left":
		switch {
		case m.MainFocus == FocusPayloadMenu && m.PayloadMenu > 0:
			m.PayloadMenu--
		case m.MainFocus == FocusURL && m.Cursor > 0:
			m.Cursor--
		case m.MainFocus == FocusMethod && m.SelectedMethod > 0:
			m.SelectedMethod--
		}
		return m, nil

	case "right":
		switch {
		case m.MainFocus == FocusPayloadMenu && m.PayloadMenu < 1:
			m.PayloadMenu++
		case m.MainFocus == FocusURL && m.Cursor < len(m.URL):
			m.Cursor++
		case m.MainFocus == FocusMethod && m.SelectedMethod < len(Methods)-1:
			m.SelectedMethod++
		}
		return m, nil

	case "up":
		switch m.MainFocus {
		case FocusPayloadMenu:
			m.MainFocus = FocusURL
		case FocusURL:
			m.MainFocus = FocusMethod
		}
		return m, nil

	case "down":
		switch m.MainFocus {
		case FocusMethod:
			m.MainFocus = FocusURL
		case FocusURL:
			m.MainFocus = FocusPayloadMenu
		}
		return m, nil

	case "home":
		if m.MainFocus == FocusURL {
			m.Cursor = 0
		}
		return m, nil

	case "end":
		if m.MainFocus == FocusURL {
			m.Cursor = len(m.URL)
		}
		return m, nil

	default:
		if m.MainFocus == FocusURL && len(msg.String()) == 1 {
			m.URL = m.URL[:m.Cursor] + msg.String() + m.URL[m.Cursor:]
			m.Cursor++
		}
		return m, nil
	}
}

// handleMainEnter opens the selected Body/Auth editor, or sends the request
// when focus is on the method selector or the URL bar.
func (m Model) handleMainEnter() (tea.Model, tea.Cmd) {
	if m.MainFocus == FocusPayloadMenu {
		if m.PayloadMenu == 1 {
			m.Screen = ScreenAuthEditor
			m.AuthInput.Focus()
		} else {
			m.Screen = ScreenPayloadEditor
			m.PayloadArea.Focus()
		}
		return m, nil
	}

	if m.Loading {
		return m, nil
	}

	m.Loading = true
	m.Response = nil
	return m, tea.Batch(
		http.MakeRequest(Methods[m.SelectedMethod], m.URL, m.PayloadArea.Value(), m.AuthInput.Value()),
		m.Spinner.Tick,
	)
}
