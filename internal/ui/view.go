package ui

import (
	"fmt"
	"time"
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
	response := m.renderResponse()

	help := HelpStyle.Render("Press Enter to go back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		RenderLogo(),
		// "",
		// TitleStyle.Render("🧇 Response"),
		"",
		response,
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

func (m Model) renderResponse() string {
	resp := m.Response
	if resp.Error != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", resp.Error))
	}
	statusColor := SuccessStyle
	if resp.StatusCode >= 400 {
		statusColor = ErrorStyle
	}
	header := fmt.Sprintf("Status: %s | Time: %v",
		statusColor.Render(fmt.Sprintf("%d", resp.StatusCode)),
		resp.ResponseTime.Round(time.Millisecond))
	displayBody := resp.Body
	if len(displayBody) > 500 {
		displayBody = displayBody[:500] + "\n... (truncated)"
	}
	return ResponseStyle.Render(
		fmt.Sprintf("%s\n\n%s", header, displayBody),
	)
}

