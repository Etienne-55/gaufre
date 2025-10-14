package ui

import (
	"fmt"
	"time"
	"gaufre/internal/types"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)


type HistoryItem struct {
	Method string
	URL string
	Response int
	Timestamp time.Time
}

func (h HistoryItem) FilterValue() string { return h.URL + " " + h.Method }
func (h HistoryItem) Title() string       { return h.Method + " " + h.URL }
func (h HistoryItem) Description() string {
	return "status: " + fmt.Sprintf("%d", h.Response) + " | " + h.Timestamp.Format("15:04:05")
}

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
	History []HistoryItem
	HistoryList list.Model
}

func NewModel() Model {
	defaultURL := "http://localhost:8080/api/rag/get_all_data"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	historyList.Title = "Request History"
	historyList.SetFilteringEnabled(true)
	historyList.SetShowStatusBar(false)

	return Model{
		URL:    defaultURL,
		Cursor: len(defaultURL),
		SelectedMethod: 0,
		SelectURL: false,
		ShowResponse: false,
		SelectPayload: false,
		Spinner: s,
		ShowHistory: false,
		History: []HistoryItem{},
		HistoryList: historyList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

