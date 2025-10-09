package ui

import (
	"gaufre/internal/types"
	tea "github.com/charmbracelet/bubbletea"
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
}

func NewModel() Model {
	defaultURL := "http://localhost:8080/api/rag/get_all_data"
	return Model{
		URL:    defaultURL,
		Cursor: len(defaultURL),
		SelectedMethod: 0,
		SelectURL: false,
		ShowResponse: false,
		SelectPayload: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

