package ui

import (
	"fmt"
	"time"
	"github.com/charmbracelet/lipgloss"
)

 
// func (m Model) View() string {
// 	var b strings.Builder
//
// 	logo := `
//   	  ████████████████████████████      
//   ███▓▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒███  
// ██▓▒▒▒▒▒▒▒▒▒▒░█▓▒▒▒▒▒▒▒▒█▓█▒▒▒▒▒▒▒▒▒▒▓██
// █▓▒▒▒▒▒▒▒▒▒░   ░▒██████▒░   ░▒▒▒▒▒▒▒▒▒▓█
// █▒▒▒▒▒▒▒▒▒▒▒▒░░▒▒▒▒▒▒▒▒▒▒░░░▒▒▒▒▒▒▒▒▒▒▒█
// █▒▒▒▒▒▒▒▓███████████▒▒▒▒▒▒▒▒▓███▒▒▒▒▒▒▒▓
// █▒▒▒▒▒▒█▓▒▒░▒█▒▒█░░▒█▒▒▒▒▒▒▓▓░░▒█▒▒▒▒▒▒█
// █▒▒▒▒▒▒█▒░▒▒░█▒▒█░░░▓▓▒▒▒▒▒█▒░░▒█▒▒▒▒▒▒█
// █▒▒▒▒▒▒███████░░▓██████▒▒▓███████▒▒▒▒▒▒█
// █▒▒▒▒▒▒█▓▓▓▓▓▒░░░▓▓▓▓▓▓▒░░▒▓▓▓▓▓█▒▒▒▒▒▒█
// █▒▒▒▒▒▒█▒▒░░░█▒▒█░░░░░░█▒▒█░░░░░█▒▒▒▒▒▒█
// ███▒▒▓█▒▒░▒░░█▒▒█░░░░░░█▒▒█░░░░░▒█▓▒▒███
//   ██▒▒█▒▒░▒░░█▒▒█░░░░░░█▒▒█░░░░░░█▒░██  
//   ██░▒█▓▓▓▓▒▓█▒▒█▓▒▒▒▒▓█▒▒█▓▒▒▒▒▓█▒░██  
//   ██░░░▒▒▒▒░░░░░░░░░░░░░░░░░░░░░░░░░██  
//   ██░▒█▓▓▓▓▓▓█▒░█▓▓▓▓▓▓█▒▒█▓▓▓▓▓▓█░░██  
//   ██░▒█▒▒░▒░░█▒▒█░░░░░░█▒▒█░░░░░░█▒░██  
//   █▒░▒█▒▒░▒░░█▒▒█░░░░░░█▒▒█░░░░░░█▒░██  
//   ██░▒█▒▒▒░▒░█▒▒█░░░░░░█▒▒█░░░░░░█▒░██  
//   ██░░▒██████▓░░▒██████▓░░▓██████▒░░██  
//    ██▒░░░░░░░░░░░░░░░░░░░░░░░░░░░░▒██   
//       ████████████████████████████                                     
//   `
func (m Model) View() string {

// 	logo := `
//  ██████╗  █████╗ ██╗   ██╗███████╗██████╗ ███████╗
// ██╔════╝ ██╔══██╗██║   ██║██╔════╝██╔══██╗██╔════╝
// ██║  ███╗███████║██║   ██║█████╗  ██████╔╝█████╗  
// ██║   ██║██╔══██║██║   ██║██╔══╝  ██╔══██╗██╔══╝  
// ╚██████╔╝██║  ██║╚██████╔╝██║     ██║  ██║███████╗
//  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝
// 	`

	methods := RenderButtons(m.SelectedMethod)

	// urlLabel := "URL: "
	// urlWithCursor := m.URL[:m.Cursor] + "|" + m.URL[m.Cursor:]
	// urlInput := InputStyle.Render(urlLabel + urlWithCursor)

	urlInput := RenderURLInput(m.URL, m.Cursor, m.SelectURL) 

	loading := ""
	if m.Loading {
		loading = "Loading..."
	}

	response := ""
	if m.Response != nil {
		response = m.renderResponse()
	}

	help := HelpStyle.Render("←→: choose method | Enter: send | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		// logo,
		RenderLogo(),
		"",
		methods,
		"",
		urlInput,
		"",
		loading,
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

