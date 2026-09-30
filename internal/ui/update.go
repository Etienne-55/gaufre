package ui

import (
	"fmt"
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
	// Some messages matter whatever screen is showing, so they are handled
	// before dispatching by screen.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Resizes affect every screen (history list size, response
		// viewport size, ...).
		return m.handleWindowSize(msg), nil

	case statusExpiredMsg:
		// A status message's expiry can land after its screen was left.
		if msg.seq == m.StatusSeq {
			m.Status = ""
		}
		return m, nil

	case http.ResponseMsg:
		// A response can arrive after you've moved on to the history or an
		// editor while it was in flight. Dispatching it by screen used to
		// drop it there, leaving Loading stuck and every later send ignored.
		return m.handleResponse(msg)

	case list.FilterMatchesMsg:
		// The history's filter results come back as a message, and can land
		// on any screen: a response re-filters the history from wherever
		// you are.
		var cmd tea.Cmd
		m.HistoryList, cmd = m.HistoryList.Update(msg)
		return m, cmd

	case spinner.TickMsg:
		// Likewise the spinner has to keep ticking off the main screen, but
		// should stop once the request is done rather than tick forever.
		if !m.Loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	switch m.Screen {
	case ScreenHistory:
		return m.updateHistoryList(msg)
	case ScreenResponse:
		return m.updateResponseViewport(msg)
	case ScreenTokenPicker:
		return m.updateTokenPicker(msg)
	case ScreenPayloadEditor:
		return m.updatePayloadEditor(msg)
	case ScreenAuthEditor:
		return m.updateAuthEditor(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleMainKeyPress(keyMsg)
	}
	return m, nil
}

// handleResponse records a finished request in the history and shows its
// response, from whichever screen you were on.
func (m Model) handleResponse(msg http.ResponseMsg) (Model, tea.Cmd) {
	m.Loading = false
	m.Response = msg.Response
	m.ResponseTokens = findResponseTokens(m.Response)
	m.TokenCursor = 0

	m.PayloadArea.Blur()
	m.AuthInput.Blur()
	m.Status = ""

	var cmd tea.Cmd
	if m.Response != nil && m.Response.Error == nil {
		// What was sent, not what is on screen now: you may have loaded
		// another request from the history while this one was in flight.
		m, cmd = m.recordHistory(types.HistoryItem{
			Method:    msg.Method,
			URL:       msg.URL,
			Payload:   msg.Payload,
			Response:  m.Response.StatusCode,
			Timestamp: time.Now(),
		})
	}

	width, height := m.responseViewportSize()
	m.Viewport = viewport.New(width, height)
	m.Viewport.SetContent(m.renderResponseContent(width))
	m.Screen = ScreenResponse
	return m, cmd
}

// recordHistory puts a request at the top of the history and saves it,
// reporting in the status line when it couldn't be saved.
func (m Model) recordHistory(item types.HistoryItem) (Model, tea.Cmd) {
	m.History = append([]types.HistoryItem{item}, m.History...)
	if len(m.History) > storage.MaxHistoryItems {
		m.History = m.History[:storage.MaxHistoryItems]
	}

	items := make([]list.Item, len(m.History))
	for i, h := range m.History {
		items[i] = h
	}
	// With a filter applied, SetItems empties the filtered list and returns
	// the command that refills it. Dropping it left the history blank.
	cmds := []tea.Cmd{m.HistoryList.SetItems(items)}

	var err error
	if m.HistoryLoadErr != nil {
		err = fmt.Errorf("it couldn't be read at startup: %w", m.HistoryLoadErr)
	} else {
		err = storage.SaveHistory(m.History)
	}
	if err != nil {
		var cmd tea.Cmd
		m, cmd = m.setStatus("history not saved: " + err.Error())
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) Model {
	m.Width = msg.Width
	m.Height = msg.Height
	m.HistoryList.SetSize(m.historyListSize())
	m.URLInput.Width = m.urlInputWidth()

	// Size the editors to the terminal: a fixed-size textarea either
	// overflowed small terminals or wasted most of a large one.
	editorWidth := m.editorContentWidth()
	m.PayloadArea.SetWidth(editorWidth)
	m.PayloadArea.SetHeight(m.editorBodyHeight())
	m.AuthInput.Width = editorWidth - len(authPrompt) - 1

	// The response is re-wrapped to the new width. This also covers the
	// token picker, which sits on top of the response viewer.
	if m.Response != nil {
		m.Viewport.Width, m.Viewport.Height = m.responseViewportSize()
		m.Viewport.SetContent(m.renderResponseContent(m.Viewport.Width))
	}

	return m
}
