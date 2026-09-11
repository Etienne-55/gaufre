package ui

import (
	"strings"
	"testing"
)

func TestResponseScreenFitsTerminal(t *testing.T) {
	jwt := testJWT(t, map[string]any{"sub": "alice"})
	bodies := []string{
		`{"ok":true}`,
		// One long minified line, which the frame used to re-wrap behind
		// the viewport's back and grow past the bottom of the terminal.
		`{"access_token":"` + jwt + `","data":"` + strings.Repeat("x", 600) + `"}`,
		strings.Repeat(`{"id":1,"name":"a fairly long name to fill the line"}`+"\n", 80),
		"<html><body>Not Found</body></html>",
	}

	for _, size := range [][2]int{{150, 45}, {100, 30}, {70, 20}, {50, 16}} {
		for _, body := range bodies {
			m := responseModel(t, size[0], size[1], body)
			assertFrameIntact(t, m.View(), size[0], size[1])

			m = press(m, "h")
			assertFrameIntact(t, m.View(), size[0], size[1])
		}
	}
}

func TestResponseHeaderStaysPinnedWhileScrolling(t *testing.T) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = `{"line":1}`
	}
	m := responseModel(t, 100, 30, strings.Join(lines, "\n"))

	m = press(m, "G")
	view := stripANSI(m.View())
	if !strings.Contains(view, "200 OK") {
		t.Error("status scrolled out of view")
	}
	if !strings.Contains(view, "of 100") {
		t.Errorf("expected a scroll position, got:\n%s", view)
	}

	m = press(m, "g")
	if m.Viewport.YOffset != 0 {
		t.Errorf("YOffset = %d after g, want 0", m.Viewport.YOffset)
	}
}

func TestResponseHeadersToggle(t *testing.T) {
	m := responseModel(t, 100, 30, `{"ok":true}`)
	m.Response.Headers = map[string][]string{"Content-Type": {"application/json; charset=utf-8"}, "X-Request-Id": {"abc-123"}}
	m.Viewport.SetContent(m.renderResponseContent(m.Viewport.Width))

	if strings.Contains(stripANSI(m.View()), "X-Request-Id") {
		t.Error("headers shown before toggling them on")
	}
	if !strings.Contains(stripANSI(m.View()), "application/json") {
		t.Error("status header should show the content type without its parameters")
	}

	m = press(m, "h")
	if view := stripANSI(m.View()); !strings.Contains(view, "X-Request-Id: abc-123") {
		t.Errorf("headers missing after h:\n%s", view)
	}

	m = press(m, "h")
	if strings.Contains(stripANSI(m.View()), "X-Request-Id") {
		t.Error("headers still shown after toggling them off")
	}
}

func TestStatusLine(t *testing.T) {
	tests := map[int]string{200: "200 OK", 404: "404 Not Found", 599: "599"}
	for code, want := range tests {
		if got := statusLine(code); got != want {
			t.Errorf("statusLine(%d) = %q, want %q", code, got, want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := map[int]string{0: "0 B", 1023: "1023 B", 2048: "2.0 KB", 3 * 1024 * 1024: "3.0 MB"}
	for n, want := range tests {
		if got := formatBytes(n); got != want {
			t.Errorf("formatBytes(%d) = %q, want %q", n, got, want)
		}
	}
}
