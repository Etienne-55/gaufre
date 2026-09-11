package ui

import (
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateHistoryList(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While the list's own filter input is active, every key (including
	// q/tab, which otherwise quit or navigate away) belongs to the filter
	// text and must be forwarded to the list instead of intercepted here.
	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.HistoryList.FilterState() != list.Filtering {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "tab":
			// With a filter applied, esc clears it first (the list's own
			// behaviour); a second esc leaves.
			if keyMsg.String() == "esc" && m.HistoryList.FilterState() == list.FilterApplied {
				break
			}
			m.Screen = ScreenMain
			return m, nil

		case "enter":
			// SelectedItem rather than an index into m.History: once the
			// list is filtered, its index counts filtered items only, which
			// used to load the wrong request.
			item, ok := m.HistoryList.SelectedItem().(types.HistoryItem)
			if !ok {
				return m, nil
			}

			m.URLInput.SetValue(item.URL)
			m.URLInput.CursorEnd()
			m.PayloadArea.SetValue(item.Payload)

			for i, method := range Methods {
				if method == item.Method {
					m.SelectedMethod = i
					break
				}
			}

			m.Screen = ScreenMain
			m.MainFocus = FocusURL
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.HistoryList, cmd = m.HistoryList.Update(msg)
	return m, cmd
}

func (m Model) updateResponseViewport(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, ResponseKeys.Quit):
			return m, tea.Quit

		case key.Matches(keyMsg, ResponseKeys.Back):
			m.Screen = ScreenMain
			m.Response = nil
			m.ResponseTokens = nil
			m.Status = ""
			return m, nil

		case key.Matches(keyMsg, ResponseKeys.UseToken):
			return m.openTokenPicker()

		case key.Matches(keyMsg, ResponseKeys.Headers):
			if m.Response.Error != nil {
				return m, nil
			}
			m.ShowHeaders = !m.ShowHeaders
			m.Viewport.SetContent(m.renderResponseContent(m.Viewport.Width))
			m.Viewport.GotoTop()
			return m, nil

		case key.Matches(keyMsg, ResponseKeys.Top):
			m.Viewport.GotoTop()
			return m, nil

		case key.Matches(keyMsg, ResponseKeys.Bottom):
			m.Viewport.GotoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}
