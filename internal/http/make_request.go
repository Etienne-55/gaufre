package http

import (
	"gaufre/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)


type ResponseMsg struct {
	Response *types.Response
}

func MakeRequest(method string, url string, payload string, authToken string) tea.Cmd {
	return func() tea.Msg {
		var response *types.Response

		switch method {
		case "GET":
			response = MakeGetRequest(url)
		case "POST":
			response = MakePostRequest(url, payload, authToken)
		case "PUT":
			response = MakePutRequest(url, payload)
		case "DELETE":
			response = MakeDeleteRequest(url)
		default:
			response = MakeGetRequest(url)
		}

		return ResponseMsg{Response: response}
	}
}

