package ui

import (
	"strings"
	"gaufre/internal/http"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)


func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q":
		return m, tea.Quit

	case "tab":
		m.ShowHistory = !m.ShowHistory
		return m, nil
	
	case "ctrl+v":
		if text, err := clipboard.ReadAll();
		err == nil {
		m.Payload = m.Payload[:m.PayloadCursor] + text + m.Payload[m.PayloadCursor:]
		m.PayloadCursor += len(text)
		}
		return m, nil

	case "ctrl+c":
		clipboard.WriteAll(m.Payload)
		return m, nil


	case "enter":
		if m.SelectPayload {
			m.Payload = m.Payload[:m.PayloadCursor] + "\n" + m.Payload[m.PayloadCursor:]
			m.PayloadCursor++
			return m, nil
		}

		if m.SelectAuth {
			m.AuthToken = m.AuthToken[:m.AuthTokenCursor] + "\n" + m.AuthToken[m.AuthTokenCursor:]
			m.AuthTokenCursor++
			return m, nil
		}

		if m.SelectPayloadMenu {
			m.SelectPayloadMenu = false
			if m.PayloadMenu == 0 {
				m.SelectPayload = true
			} else {
				m.SelectAuth = true
			}
			return m, nil
		}

		if m.SelectURL && !m.Loading {
			m.Loading = true
			m.Response = nil
			m.ShowResponse = false
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			return m, tea.Batch(
				http.MakeRequest(methods[m.SelectedMethod], m.URL, m.Payload),
				m.Spinner.Tick,
			)
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
	} else if m.SelectAuth && m.AuthTokenCursor > 0 {
		m.AuthTokenCursor--
	} else if m.SelectPayloadMenu && m.PayloadMenu > 0 {
		m.PayloadMenu--
	} else if m.SelectURL && m.Cursor > 0 {
		m.Cursor--
	} else if !m.SelectURL && !m.SelectPayload && !m.SelectAuth && !m.SelectPayloadMenu && m.SelectedMethod > 0 {
		m.SelectedMethod--
	}
	return m, nil

case "right":
	if m.SelectPayload && m.PayloadCursor < len(m.Payload) {
		m.PayloadCursor++
	} else if m.SelectAuth && m.AuthTokenCursor < len(m.AuthToken) {
		m.AuthTokenCursor++
	} else if m.SelectPayloadMenu && m.PayloadMenu < 1 {
		m.PayloadMenu++
	} else if m.SelectURL && m.Cursor < len(m.URL) {
		m.Cursor++
	} else if !m.SelectURL && !m.SelectPayload && !m.SelectAuth && !m.SelectPayloadMenu && m.SelectedMethod < 3 {
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
	} else if m.SelectAuth {
		lastNewline := strings.LastIndex(m.AuthToken[:m.AuthTokenCursor], "\n")
		if lastNewline != -1 {
			prevNewline := strings.LastIndex(m.AuthToken[:lastNewline], "\n")
			colPos := m.AuthTokenCursor - lastNewline - 1
			if prevNewline == -1 {
				m.AuthTokenCursor = min(colPos, lastNewline)
			} else {
				lineLen := lastNewline - prevNewline - 1
				m.AuthTokenCursor = prevNewline + 1 + min(colPos, lineLen)
			}
		}
	} else if m.SelectPayloadMenu {
		m.SelectPayloadMenu = false
		m.SelectURL = true
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
		} else if m.SelectAuth {
			lastNewline := strings.LastIndex(m.AuthToken[:m.AuthTokenCursor], "\n")
			colPos := m.AuthTokenCursor - lastNewline - 1
			nextNewline := strings.Index(m.AuthToken[m.AuthTokenCursor:], "\n")
			if nextNewline != -1 {
				nextNewline += m.AuthTokenCursor
				nextNextNewline := strings.Index(m.AuthToken[nextNewline+1:], "\n")
				if nextNextNewline == -1 {
					lineLen := len(m.AuthToken) - nextNewline - 1
					m.AuthTokenCursor = nextNewline + 1 + min(colPos, lineLen)
				} else {
					nextNextNewline += nextNewline + 1
					lineLen := nextNextNewline - nextNewline - 1
					m.AuthTokenCursor = nextNewline + 1 + min(colPos, lineLen)
				}
			}
		} else if m.SelectPayloadMenu {
			return m, nil
		} else if m.SelectURL {
			m.SelectURL = false
			m.SelectPayloadMenu = true
		} else if !m.SelectURL && !m.SelectPayloadMenu {
			m.SelectURL = true
		}
		return m, nil
		
	case "esc":
		if m.SelectPayload {
			m.SelectPayload = false
			m.SelectURL = true
			return m, nil
		}

		if m.SelectAuth {
			m.SelectAuth = false
			m.SelectPayloadMenu = true
			return m, nil
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

