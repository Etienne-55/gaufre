package ui

import (
	"gaufre/internal/http"
	"gaufre/internal/storage"
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)


func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.ShowHistory {
		return m.updateHistoryList(msg)
	}

	if m.ShowResponse {
		return m.updateResponseViewport(msg)
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

		if m.ShowResponse {
			headerHeight := 4
			footerHeight := 3
			verticalMargin := headerHeight + footerHeight

			if !m.ViewportReady {
				m.Viewport = viewport.New(msg.Width-10, msg.Height-verticalMargin)
				m.ViewportReady = true
				if m.Response != nil {
					m.Viewport.SetContent(m.renderResponse())
				}
			} else {
				m.Viewport.Width = msg.Width - 10
				m.Viewport.Height = msg.Height - verticalMargin
			}
		}
		return m, nil

	case http.ResponseMsg:
		m.Loading = false
		m.Response = msg.Response
		if m.Response != nil && m.Response.Error == nil {
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			histItem := types.HistoryItem{
				Method: methods[m.SelectedMethod],
				URL: m.URL,
				Payload: m.Payload,
			}
			m.History = append([]types.HistoryItem{histItem}, m.History...)
	
			storage.SaveHistory(m.History)
			items := make([]list.Item, len(m.History))
			for i, h := range m.History {
				items[i] = h
			}
			m.HistoryList.SetItems(items)
		}
		m.Viewport = viewport.New(200, m.Height-10)
		m.Viewport.SetContent(m.renderResponseContent())
		m.ViewportReady = true
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

