package http

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"gaufre/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

// ResponseMsg reports a finished request along with what was sent: the
// method, URL and body on screen may have changed while it was in flight,
// so they can't be read back from the UI when it lands.
type ResponseMsg struct {
	Method   string
	URL      string
	Payload  string
	Response *types.Response
}

// client is shared across requests so connections can be reused instead of
// paying a fresh dial + TLS handshake on every send.
var client = &http.Client{
	Timeout: 10 * time.Second,
}

// MakeRequest returns a tea.Cmd that performs an HTTP request in the
// background and reports the outcome as a ResponseMsg.
func MakeRequest(method string, url string, payload string, authToken string) tea.Cmd {
	return func() tea.Msg {
		return ResponseMsg{
			Method:   method,
			URL:      url,
			Payload:  payload,
			Response: Do(method, url, payload, authToken),
		}
	}
}

// Do performs a single HTTP request and returns the result synchronously.
// It is the one place that builds and sends requests for every verb the UI
// exposes (GET/POST/PUT/DELETE), so the body and the Bearer token are
// applied consistently regardless of method.
func Do(method, url, body, authToken string) *types.Response {
	start := time.Now()

	// GET requests never carry a body, matching the previous per-method
	// implementations and avoiding surprises with servers/proxies that
	// reject or strip a GET body.
	var bodyReader io.Reader
	if body != "" && method != "GET" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return &types.Response{Error: err, ResponseTime: time.Since(start)}
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(authToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return &types.Response{Error: err, ResponseTime: time.Since(start)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.Response{
			StatusCode:   resp.StatusCode,
			Error:        err,
			ResponseTime: time.Since(start),
		}
	}

	return &types.Response{
		Body:         string(respBody),
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header,
		ResponseTime: time.Since(start),
	}
}
