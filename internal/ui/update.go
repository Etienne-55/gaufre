package ui

import (
	"strings"
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


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c", "q":
		return m, tea.Quit
	
	case "enter":

		if m.SelectPayload {
			m.Payload = m.Payload[:m.PayloadCursor] + "\n" + m.Payload[m.PayloadCursor:]
			m.PayloadCursor++
			return m, nil
		}

		if m.SelectPayload {
			m.SelectPayload = false
			m.SelectURL = true
			return m, nil
		}

		if m.SelectURL && !m.Loading {
			m.Loading = true
			m.Response = nil
			m.ShowResponse = false
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			return m, 
			makeRequest(methods[m.SelectedMethod], m.URL, m.Payload)
		}
		return m, nil

	case "backspace":
		if m.SelectPayload && len(m.Payload) > 0 && m.PayloadCursor > 0 {
			m.Payload = m.Payload[:m.PayloadCursor-1] + m.Payload[m.PayloadCursor:]
			m.PayloadCursor--
		} else if  m.SelectURL && len(m.URL) > 0 && m.Cursor > 0 {
			m.URL = m.URL[:m.Cursor-1] + m.URL[m.Cursor:]
			m.Cursor--
		}
		return m, nil

	case "left":
		if m.SelectPayload && m.PayloadCursor > 0 {
			m.PayloadCursor--
		} else if m.SelectURL && m.Cursor > 0 {
			m.Cursor--
		} else if !m.SelectURL && !m.SelectPayload && m.SelectedMethod > 0 {
			m.SelectedMethod--
		}
		return m, nil

	case "right":
		if m.SelectPayload && m.PayloadCursor < len(m.Payload) {
			m.PayloadCursor++
		} else if m.SelectURL && m.Cursor < len(m.URL) {
			m.Cursor++
		} else if !m.SelectURL && !m.SelectPayload && m.SelectedMethod < 3 {
			m.SelectedMethod++
		}
		return m, nil

	case "up":
		if m.SelectPayload {
			lastNewline := strings.LastIndex(m.Payload[:m.PayloadCursor], "\n")
			if lastNewline != -1 {
				prevNewline := strings.LastIndex(m.Payload[:lastNewline], "\n")
				colPos := m.PayloadCursor - lastNewline - 1
				if prevNewline == -1 {
					m.PayloadCursor = min(colPos, lastNewline)
				} else {
					lineLen := lastNewline - prevNewline - 1
					m.PayloadCursor = prevNewline + 1 + min(colPos, lineLen)
				}
			}
		} else if m.SelectURL {
			m.SelectURL = false
		}
		return m, nil

	case "down":
		if m.SelectPayload {
			lastNewline := strings.LastIndex(m.Payload[:m.PayloadCursor], "\n")
			colPos := m.PayloadCursor - lastNewline - 1
			nextNewline := strings.Index(m.Payload[m.PayloadCursor:], "\n")
			if nextNewline != -1 {
				nextNewline += m.PayloadCursor
				nextNextNewline := strings.Index(m.Payload[nextNewline+1:], "\n")
				if nextNextNewline == -1 {
					lineLen := len(m.Payload) - nextNewline - 1
					m.PayloadCursor = nextNewline + 1 + min(colPos, lineLen)
				} else {
					nextNextNewline += nextNewline + 1
					lineLen := nextNextNewline - nextNewline - 1
					m.PayloadCursor = nextNewline + 1 + min(colPos, lineLen)
				}
			}
		} else if !m.SelectURL {
			m.SelectURL = true
		} else if  m.SelectURL && (m.SelectedMethod == 1 || m.SelectedMethod == 2){
			m.SelectURL = false
			m.SelectPayload = true
		}
		return m, nil
		
	case "esc":
		if m.SelectPayload {
			m.SelectPayload = false
			m.SelectURL = true
		}

		if m.ShowResponse {
			m.ShowResponse = false
			m.Response = nil
			return m, nil
		}
		return m, nil

	case "home":
		m.Cursor = 0
		return m, nil

	case "end":
		m.Cursor = len(m.URL) 
		return m, nil

	default:
		if m.SelectPayload && len(msg.String()) == 1 {
			m.Payload = m.Payload[:m.PayloadCursor] + msg.String() + m.Payload[m.PayloadCursor:]
			m.PayloadCursor++
			return m, nil
		}
		if m.SelectURL && len(msg.String()) == 1 {
			m.URL = m.URL[:m.Cursor] + msg.String() + m.URL[m.Cursor:]
			m.Cursor++

			return m, nil
		}
	}

	return m, nil
}

func makeRequest(method string, url string, payload string) tea.Cmd {
	return func() tea.Msg {
		var response *types.Response

		switch method {
		case "GET":
			response = http.MakeGetRequest(url)
		case "POST":
			response = http.MakePostRequest(url, payload)
		// case "PUT":
		// 	response = http.MakePutRequest(url, payload)
		// case "DELETE":
		// 	response = http.MakeDeleteRequest(url)
		default:
			response = http.MakeGetRequest(url)
		}

		return ResponseMsg{Response: response}
	}
}

