package ui

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func testJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	enc := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return enc(map[string]any{"alg": "HS256", "typ": "JWT"}) + "." + enc(claims) + ".c2lnbmF0dXJl"
}

func sizedModel(width, height int) Model {
	m := NewModel()
	m.URL = "http://localhost:3000/api/v1/users"
	m.SelectedMethod = 1
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(Model)
}

// assertFrameIntact is the regression guard for the editors' layout: a child
// wider than its container makes lipgloss re-wrap the already-drawn box,
// which is what used to shred the token editor into stray border characters.
func assertFrameIntact(t *testing.T, out string, width, height int) {
	t.Helper()

	lines := strings.Split(out, "\n")
	if len(lines) > height {
		t.Errorf("rendered %d lines, terminal height is %d", len(lines), height)
	}

	for i, line := range lines {
		if w := len([]rune(stripANSI(line))); w > width {
			t.Errorf("line %d is %d cells wide, terminal width is %d:\n%s", i, w, width, stripANSI(line))
		}
	}

	var frameWidths []int
	for _, line := range lines {
		plain := strings.TrimRight(stripANSI(line), " ")
		if strings.Contains(plain, "╭") || strings.Contains(plain, "╰") {
			frameWidths = append(frameWidths, len([]rune(plain)))
		}
	}
	if len(frameWidths) != 2 {
		t.Fatalf("expected exactly 2 frame corner rows (top and bottom), got %d", len(frameWidths))
	}
	if frameWidths[0] != frameWidths[1] {
		t.Errorf("top and bottom frame rows differ in width: %d vs %d", frameWidths[0], frameWidths[1])
	}

	// Every row between the corners must carry both side borders at the
	// same columns, i.e. the box is a rectangle.
	want := frameWidths[0]
	for _, line := range lines {
		plain := strings.TrimRight(stripANSI(line), " ")
		if !strings.Contains(plain, "│") {
			continue
		}
		if got := len([]rune(plain)); got != want {
			t.Errorf("frame row is %d cells wide, corners are %d:\n%s", got, want, plain)
		}
	}
}

func TestEditorScreensFitTerminal(t *testing.T) {
	longToken := testJWT(t, map[string]any{
		"sub": "1234567890",
		"iss": "https://auth.example.com",
		"aud": []any{"api://gaufre"},
		"iat": float64(time.Now().Add(-time.Hour).Unix()),
		"exp": float64(time.Now().Add(42 * time.Minute).Unix()),
	})

	sizes := [][2]int{{120, 40}, {90, 28}, {64, 18}, {40, 12}}
	bodies := []string{
		"",
		"{\n  \"name\": \"test\",\n  \"count\": 12\n}",
		"{\n  \"broken\": \n}",
		strings.Repeat("{\"very\":\"long line that keeps going and going and going\"}", 4),
	}
	tokens := []string{
		"",
		longToken,
		"sk_live_51H8xYzAbCdEf+GhIjKlMnOp/QrStUvWx=",
		strings.Repeat("abcdefghij", 40),
	}

	for _, size := range sizes {
		width, height := size[0], size[1]

		for _, body := range bodies {
			m := sizedModel(width, height)
			m.Screen = ScreenPayloadEditor
			m.PayloadArea.Focus()
			m.PayloadArea.SetValue(body)
			assertFrameIntact(t, m.View(), width, height)
		}

		for _, token := range tokens {
			m := sizedModel(width, height)
			m.Screen = ScreenAuthEditor
			m.AuthInput.Focus()
			m.AuthInput.SetValue(token)
			assertFrameIntact(t, m.View(), width, height)

			// Same again with a transient status message displayed.
			m.Status = "JWT pasted and decoded"
			assertFrameIntact(t, m.View(), width, height)
		}
	}
}

func TestEditorScreensShowContextAndState(t *testing.T) {
	m := sizedModel(100, 30)
	m.Screen = ScreenPayloadEditor
	m.PayloadArea.SetValue("{\"a\": 1}")

	view := stripANSI(m.View())
	for _, want := range []string{"Request Body", "POST http://localhost:3000/api/v1/users", "valid JSON", "ctrl+f", "esc"} {
		if !strings.Contains(view, want) {
			t.Errorf("body editor view missing %q", want)
		}
	}

	m.PayloadArea.SetValue("{\"a\": }")
	if view := stripANSI(m.View()); !strings.Contains(view, "line 1") {
		t.Errorf("invalid body should report a line number, got:\n%s", view)
	}

	m.Screen = ScreenAuthEditor
	m.AuthInput.SetValue(testJWT(t, map[string]any{"sub": "user-42", "iss": "https://issuer.test"}))
	view = stripANSI(m.View())
	for _, want := range []string{"Authorization", "Bearer", "JWT", "HS256", "user-42", "https://issuer.test"} {
		if !strings.Contains(view, want) {
			t.Errorf("auth editor view missing %q", want)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
		{"hello", -1, ""},
	}

	for _, tt := range tests {
		if got := truncateString(tt.in, tt.max); got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
		}
	}
}
