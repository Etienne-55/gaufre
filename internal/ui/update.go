package ui

import (
	"time"

	"gaufre/internal/http"
	"gaufre/internal/storage"
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Window resizes affect every screen (history list size, response
	// viewport size, ...), so handle them before dispatching by screen.
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		return m.handleWindowSize(sizeMsg), nil
	}

	// A status message's expiry can land after its editor was left, so it
	// is handled globally rather than per screen.
	if expired, ok := msg.(statusExpiredMsg); ok {
		if expired.seq == m.StatusSeq {
			m.Status = ""
		}
		return m, nil
	}

	switch m.Screen {
	case ScreenHistory:
		return m.updateHistoryList(msg)
	case ScreenResponse:
		return m.updateResponseViewport(msg)
	case ScreenPayloadEditor:
		return m.updatePayloadEditor(msg)
	case ScreenAuthEditor:
		return m.updateAuthEditor(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleMainKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case http.ResponseMsg:
		m.Loading = false
		m.Response = msg.Response
		if m.Response != nil && m.Response.Error == nil {
			histItem := types.HistoryItem{
				Method:    Methods[m.SelectedMethod],
				URL:       m.URL,
				Payload:   m.PayloadArea.Value(),
				Response:  m.Response.StatusCode,
				Timestamp: time.Now(),
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
		m.Screen = ScreenResponse
		return m, nil
	}
	return m, nil
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) Model {
	m.Width = msg.Width
	m.Height = msg.Height
	m.HistoryList.SetSize(msg.Width-4, msg.Height-10)

	// Size the editors to the terminal: a fixed-size textarea either
	// overflowed small terminals or wasted most of a large one.
	editorWidth := m.editorContentWidth()
	m.PayloadArea.SetWidth(editorWidth)
	m.PayloadArea.SetHeight(m.editorBodyHeight())
	m.AuthInput.Width = editorWidth - len(authPrompt) - 1

	if m.Screen == ScreenResponse {
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

	return m
}
