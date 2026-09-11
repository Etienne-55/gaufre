package ui

import (
	"strings"

	"gaufre/internal/http"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMainKeyPress handles input for the main screen only (method
// selector, URL bar, Body/Auth menu). Payload/Auth editing, history and the
// response viewer each own their input handling in their own file, since
// each is really a separate screen with its own key semantics.
func (m Model) handleMainKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, MainKeys.ForceQuit):
		return m, tea.Quit

	case key.Matches(msg, MainKeys.History):
		m.Screen = ScreenHistory
		return m, nil

	case key.Matches(msg, MainKeys.Enter):
		return m.handleMainEnter()

	case key.Matches(msg, MainKeys.Up):
		switch m.MainFocus {
		case FocusPayloadMenu:
			m.MainFocus = FocusURL
		case FocusURL:
			m.MainFocus = FocusMethod
		}
		return m, nil

	case key.Matches(msg, MainKeys.Down):
		switch m.MainFocus {
		case FocusMethod:
			m.MainFocus = FocusURL
		case FocusURL:
			m.MainFocus = FocusPayloadMenu
		}
		return m, nil
	}

	// While the URL has focus every other key is text: "q" has to type a q
	// (think "?q=search"), not quit.
	if m.MainFocus == FocusURL {
		return m.updateURLInput(msg)
	}

	switch {
	case key.Matches(msg, MainKeys.Quit):
		return m, tea.Quit

	case key.Matches(msg, MainKeys.Choose):
		step := 1
		if msg.String() == "left" {
			step = -1
		}
		switch m.MainFocus {
		case FocusMethod:
			m.SelectedMethod = clampInt(m.SelectedMethod+step, 0, len(Methods)-1)
		case FocusPayloadMenu:
			m.PayloadMenu = clampInt(m.PayloadMenu+step, 0, 1)
		}
	}

	return m, nil
}

// updateURLInput forwards a key to the URL field. Pasting is handled here
// rather than by the textinput so a URL copied with a trailing newline, or
// wrapped across lines, doesn't end up with spaces in it.
func (m Model) updateURLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+v" {
		text, err := clipboard.ReadAll()
		if err != nil {
			return m, nil
		}
		text = strings.Join(strings.Fields(text), "")

		value := []rune(m.URLInput.Value())
		pos := m.URLInput.Position()
		m.URLInput.SetValue(string(value[:pos]) + text + string(value[pos:]))
		m.URLInput.SetCursor(pos + len([]rune(text)))
		return m, nil
	}

	var cmd tea.Cmd
	m.URLInput, cmd = m.URLInput.Update(msg)
	return m, cmd
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
		http.MakeRequest(Methods[m.SelectedMethod], m.URLInput.Value(), m.PayloadArea.Value(), m.AuthInput.Value()),
		m.Spinner.Tick,
	)
}
