package ui

import (
	"gaufre/internal/types"
	"gaufre/internal/storage"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
)


type Model struct {
	URL      string
	Response *types.Response
	Loading  bool
	Cursor   int
	Width int
	Height int
	SelectedMethod int
	SelectedOperation string
	SelectURL bool
	ShowResponse bool
	Payload string
	PayloadCursor int
	SelectPayload bool
	Spinner spinner.Model
	ShowHistory bool
	History []types.HistoryItem
	HistoryList list.Model
	Viewport viewport.Model
	ViewportReady bool
}

func NewModel() Model {
	defaultURL := "http://localhost:8080/events"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	historyList.Title = "Request History"
	historyList.SetFilteringEnabled(true)
	historyList.SetShowStatusBar(false)

	history := []types.HistoryItem{}
	if loaded, err := storage.LoadHistory(); err == nil {
		history = loaded
		items := make([]list.Item, len(history))
		for i, h := range history {
			items[i] = h
		}
		historyList.SetItems(items)
	}

	return Model{
		URL:    defaultURL,
		Cursor: len(defaultURL),
		SelectedMethod: 0,
		SelectURL: false,
		ShowResponse: false,
		SelectPayload: false,
		Spinner: s,
		ShowHistory: false,
		History: history,
		HistoryList: historyList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

