package ui

import (
	"gaufre/internal/http"
	"gaufre/internal/types"
	"gaufre/internal/storage"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)


func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.ShowHistory {
		return m.updateHistoryList(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg: 
		m.Width = msg.Width
		m.Height = msg.Height
		m.HistoryList.SetSize(msg.Width-4, msg.Height-10)
		return m, nil 

	case http.ResponseMsg:
		m.Loading = false
		m.Response = msg.Response
		if m.Response != nil && m.Response.Error == nil {
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			histItem := types.HistoryItem{
				Method: methods[m.SelectedMethod],
				URL: m.URL,
			}
			m.History = append([]types.HistoryItem{histItem}, m.History...)
	
			storage.SaveHistory(m.History)
			items := make([]list.Item, len(m.History))
			for i, h := range m.History {
				items[i] = h
			}
			m.HistoryList.SetItems(items)
		}
		m.ShowResponse = true
		return m, nil
	}
	return m, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

