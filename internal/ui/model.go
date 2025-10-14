package ui

import (
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

func NewModel() Model {
	defaultURL := "http://localhost:8080/api/rag/get_all_data"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		URL:    defaultURL,
		Cursor: len(defaultURL),
		SelectedMethod: 0,
		SelectURL: false,
		ShowResponse: false,
		SelectPayload: false,
		Spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

