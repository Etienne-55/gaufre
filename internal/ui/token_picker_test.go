package ui

import (
	"errors"
	"strings"
	"testing"

	"gaufre/internal/http"
	"gaufre/internal/types"
)

var errTest = errors.New("boom")

// responseModel is a model showing body as a received response, the way it
// looks right after a request comes back.
func responseModel(t *testing.T, width, height int, body string) Model {
	t.Helper()
	m := sizedModel(width, height)
	next, _ := m.Update(http.ResponseMsg{Response: &types.Response{Body: body, StatusCode: 200}})
	m = next.(Model)
	if m.Screen != ScreenResponse {
		t.Fatalf("Screen = %v, want ScreenResponse", m.Screen)
	}
	return m
}

func TestResponseSingleTokenGoesStraightToAuth(t *testing.T) {
	jwt := testJWT(t, map[string]any{"sub": "alice"})
	m := responseModel(t, 100, 30, `{"access_token":"`+jwt+`","expires_in":3600}`)

	m = press(m, "t")

	if m.Screen != ScreenResponse {
		t.Errorf("Screen = %v, want to stay on the response", m.Screen)
	}
	if got := m.AuthInput.Value(); got != jwt {
		t.Errorf("auth token = %q, want the response's access_token", got)
	}
	if !strings.Contains(m.Status, "access_token") {
		t.Errorf("status = %q, want it to name where the token came from", m.Status)
	}

	// The token survives leaving the response, so the next request uses it.
	m = press(m, "esc")
	if got := m.AuthInput.Value(); got != jwt {
		t.Errorf("auth token after esc = %q, want it kept", got)
	}
}

func TestResponseSeveralTokensOpensPicker(t *testing.T) {
	access := testJWT(t, map[string]any{"sub": "alice"})
	refresh := testJWT(t, map[string]any{"sub": "alice", "kind": "refresh"})
	m := responseModel(t, 100, 30, `{"access_token":"`+access+`","refresh_token":"`+refresh+`"}`)

	m = press(m, "t")
	if m.Screen != ScreenTokenPicker {
		t.Fatalf("Screen = %v, want ScreenTokenPicker", m.Screen)
	}

	out := m.View()
	assertFrameIntact(t, out, 100, 30)
	for _, want := range []string{"Tokens in response (2)", "access_token", "refresh_token"} {
		if !strings.Contains(out, want) {
			t.Errorf("picker view is missing %q", want)
		}
	}

	m = press(m, "down", "enter")
	if m.Screen != ScreenResponse {
		t.Errorf("Screen = %v, want back on the response", m.Screen)
	}
	if got := m.AuthInput.Value(); got != refresh {
		t.Errorf("auth token = %q, want the selected refresh_token", got)
	}
}

func TestTokenPickerEscKeepsCurrentToken(t *testing.T) {
	m := responseModel(t, 100, 30, `{"access_token":"first-token-value","id_token":"second-token-value"}`)
	m.AuthInput.SetValue("existing")

	m = press(m, "t", "esc")
	if m.Screen != ScreenResponse {
		t.Errorf("Screen = %v, want ScreenResponse", m.Screen)
	}
	if got := m.AuthInput.Value(); got != "existing" {
		t.Errorf("auth token = %q, esc must not change it", got)
	}
}

func TestResponseWithoutTokenSaysSo(t *testing.T) {
	m := responseModel(t, 100, 30, `{"id":1}`)

	m = press(m, "t")
	if m.Screen != ScreenResponse {
		t.Errorf("Screen = %v, want ScreenResponse", m.Screen)
	}
	if m.Status == "" {
		t.Error("expected a status explaining no token was found")
	}
	if strings.Contains(m.View(), "use token") || strings.Contains(m.View(), "pick token") {
		t.Error("help advertises the token shortcut for a response without a token")
	}
}

func TestTokenPickerFitsSmallTerminals(t *testing.T) {
	body := `{"sessions":[`
	for i := 0; i < 20; i++ {
		if i > 0 {
			body += ","
		}
		body += `{"token":"token-value-number-` + string(rune('a'+i)) + `"}`
	}
	body += `]}`

	for _, size := range [][2]int{{60, 16}, {100, 30}, {200, 60}} {
		m := responseModel(t, size[0], size[1], body)
		m = press(m, "t")
		// Move to the bottom so the list has to scroll.
		for i := 0; i < 25; i++ {
			m = press(m, "down")
		}
		out := m.View()
		assertFrameIntact(t, out, size[0], size[1])
		if !strings.Contains(out, "sessions[19].token") {
			t.Errorf("%dx%d: selected last entry is scrolled out of view", size[0], size[1])
		}
	}
}

func TestVisibleWindow(t *testing.T) {
	tests := []struct {
		n, cursor, capacity int
		first, last         int
	}{
		{5, 0, -1, 0, 5},
		{5, 4, 10, 0, 5},
		{10, 0, 4, 0, 4},
		{10, 5, 4, 3, 7},
		{10, 9, 4, 6, 10},
	}
	for _, tt := range tests {
		first, last := visibleWindow(tt.n, tt.cursor, tt.capacity)
		if first != tt.first || last != tt.last {
			t.Errorf("visibleWindow(%d, %d, %d) = [%d, %d), want [%d, %d)",
				tt.n, tt.cursor, tt.capacity, first, last, tt.first, tt.last)
		}
	}
}

func TestTruncateMiddleKeepsBothEnds(t *testing.T) {
	if got := truncateMiddle("abcdefghijklmnopqrstuvwxyz", 11); got != "abcde…vwxyz" {
		t.Errorf("truncateMiddle = %q, want abcde…vwxyz", got)
	}
	if got := truncateMiddle("short", 11); got != "short" {
		t.Errorf("truncateMiddle changed a string that fits: %q", got)
	}
}

// Only the selected candidate carries the selection bar, as in the history.
func TestTokenPickerMarksOnlyTheSelectedToken(t *testing.T) {
	m := responseModel(t, 100, 30, `{"access_token":"first-token-value","id_token":"second-token-value"}`)
	m = press(m, "t", "down")

	var barred []string
	for _, line := range strings.Split(stripANSI(m.View()), "\n") {
		if inner := strings.TrimSpace(strings.Trim(strings.TrimSpace(line), "│")); strings.HasPrefix(inner, "│ ") {
			barred = append(barred, strings.TrimPrefix(inner, "│ "))
		}
	}
	if len(barred) != 3 || barred[0] != "id_token" || barred[2] != "second-token-value" {
		t.Errorf("selection bar rows = %q, want the id_token card's three rows", barred)
	}
}
