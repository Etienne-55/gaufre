package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)


func (m Model) updateHistoryList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "esc", "1":
			m.ShowHistory = false
			return m, nil

		case "enter":
			if len(m.History) > 0 {
				selectedIdx := m.HistoryList.Index()
				if selectedIdx >= 0 && selectedIdx < len(m.History) {
					item := m.History[selectedIdx]
					m.URL = item.URL
					m.Cursor = len(m.URL)
					m.ShowHistory = false
					m.SelectURL = true
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

