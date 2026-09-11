package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginBottom(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	ResponseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

// focusedButtonStyle is the selected button while its row has focus: filled,
// and outlined in the same color so it keeps the size of the other buttons
// (without the border, the whole row shifted by two columns as focus
// entered and left it).
var focusedButtonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#874BFD")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Bold(true).
	Padding(0, 2)

var idleButtonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#874BFD")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(0, 2)

func RenderMethodButtons(selected int, focused bool) string {
	var buttons []string

	for i, method := range Methods {
		switch {
		case i == selected && focused:
			buttons = append(buttons, focusedButtonStyle.Render(method))
		case i == selected:
			buttons = append(buttons, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")).
				Bold(true).
				Padding(0, 2).
				Render(method))
		default:
			buttons = append(buttons, idleButtonStyle.Render(method))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
}

func RenderPayloadButtons(selected int, focused bool) string {
	var buttons []string

	for i, option := range []string{"Body", "Auth"} {
		if focused && i == selected {
			buttons = append(buttons, focusedButtonStyle.Render(option))
		} else {
			buttons = append(buttons, idleButtonStyle.Render(option))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
}
