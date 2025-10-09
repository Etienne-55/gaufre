package ui

import (
	"github.com/charmbracelet/lipgloss"
)
 
func (m Model) View() string {

	if m.ShowResponse && m.Response != nil {
		return m.renderResponseScreen()
	}

	methods := RenderButtons(m.SelectedMethod)
	urlInput := RenderURLInput(m.URL, m.Cursor, m.SelectURL) 
	loading := ""
	if m.Loading {
		loading = "Loading..."
	}

	help := HelpStyle.Render("←→: choose method | Enter: send | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		RenderLogo(),
		"",
		methods,
		"",
		urlInput,
		"",
		loading,
		"",
		help,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(2)

	boxedContent := boxStyle.Render(content)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderResponseScreen() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	availableWidth := m.Width - 10
	availableHeight := m.Height - 8

	response := m.renderResponse()

	styledResponse := lipgloss.NewStyle().
		Width(availableWidth).
		Height(availableHeight).
		AlignVertical(lipgloss.Top).
		Render(response)


	help := HelpStyle.Render("Press Enter to go back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		styledResponse,
		"",
		help,
	)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

