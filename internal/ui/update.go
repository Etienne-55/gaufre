package ui

import (
	"gaufre/internal/http"
	"gaufre/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)


type ResponseMsg struct {
	Response *types.Response
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg: 
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil 

	case ResponseMsg:
		m.Loading = false
		m.Response = msg.Response
		m.ShowResponse = true
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c", "q":
		return m, tea.Quit
	
	case "enter":
		if m.ShowResponse {
			m.ShowResponse = false
			m.Response = nil
			return m, nil
		}

		if m.SelectURL && !m.Loading {
			m.Loading = true
			m.Response = nil
			m.ShowResponse = false
			return m, makeRequest(m.URL)
		}

	case "backspace":
		if  m.SelectURL && len(m.URL) > 0 && m.Cursor > 0 {
			m.URL = m.URL[:m.Cursor-1] + m.URL[m.Cursor:]
			m.Cursor--
		}

	case "left":
		if !m.SelectURL && m.SelectedMethod > 0 {
			m.SelectedMethod--
		}

	case "right":
		if !m.SelectURL && m.SelectedMethod < 3 {
			m.SelectedMethod++
		}

	case "up":
		if m.SelectURL {
			m.SelectURL = false
		}

	case "down":
		if !m.SelectURL {
			m.SelectURL = true
		}

	case "home":
		m.Cursor = 0

	case "end":
		m.Cursor = len(m.URL) 

	default:
		if m.SelectURL && len(msg.String()) == 1 {
			m.URL = m.URL[:m.Cursor] + msg.String() + m.URL[m.Cursor:]
			m.Cursor++
		}
	}

	return m, nil
}

func makeRequest(url string) tea.Cmd {
	return func() tea.Msg {
		response := http.MakeGetRequest(url)
		return ResponseMsg{Response: response}
	}
}

