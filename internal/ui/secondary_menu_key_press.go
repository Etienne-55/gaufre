package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateHistoryList(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While the list's own filter input is active, every key (including
	// q/tab, which otherwise quit or navigate away) belongs to the filter
	// text and must be forwarded to the list instead of intercepted here.
	if key, ok := msg.(tea.KeyMsg); ok && m.HistoryList.FilterState() != list.Filtering {
		switch key.String() {
		case "q":
			return m, tea.Quit

		case "esc", "tab":
			m.Screen = ScreenMain
			return m, nil

		case "enter":
			if len(m.History) > 0 {
				selectedIdx := m.HistoryList.Index()
				if selectedIdx >= 0 && selectedIdx < len(m.History) {
					item := m.History[selectedIdx]
					m.URL = item.URL
					m.Cursor = len(m.URL)
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
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.HistoryList, cmd = m.HistoryList.Update(msg)
	return m, cmd
}

func (m Model) updateResponseViewport(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc":
			m.Screen = ScreenMain
			m.Response = nil
			m.ResponseTokens = nil
			m.Status = ""
			m.ViewportReady = false
			return m, nil
		case "t":
			return m.openTokenPicker()
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}
