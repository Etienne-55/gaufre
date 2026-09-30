package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// requestContext is the "POST http://localhost:3000/users" line shown in the
// top-right of both editors, so you can see what you're editing for.
func (m Model) requestContext() string {
	return Methods[m.SelectedMethod] + " " + m.URLInput.Value()
}

func (m Model) renderPayloadScreen() string {
	width := m.editorContentWidth()

	help := m.Help
	help.Width = width

	return m.renderEditorFrame(width,
		renderTitleBar(width, "Request Body", m.requestContext()),
		"",
		m.PayloadArea.View(),
		"",
		m.renderBodyStatus(width),
		help.View(PayloadKeys),
	)
}

// renderBodyStatus shows either the transient action message (copied,
// formatted, ...) or live JSON validity for what's currently in the editor.
func (m Model) renderBodyStatus(width int) string {
	if m.Status != "" {
		return EditorOKStyle.Render(truncateString(m.Status, width))
	}

	status := inspectBody(m.PayloadArea.Value())
	meta := EditorMetaStyle.Render(fmt.Sprintf("%d lines · %d bytes", status.Lines, status.Bytes))

	var state string
	switch {
	case status.Empty:
		state = EditorMetaStyle.Render("empty body")
	case status.Valid:
		state = EditorOKStyle.Render("✓ valid JSON")
	default:
		state = EditorBadStyle.Render("✗ " + status.Problem)
	}

	// Keep the two halves on one line when they fit, otherwise drop the
	// counts rather than wrapping the frame.
	combined := state + EditorMetaStyle.Render("  ·  ") + meta
	if lipgloss.Width(combined) <= width {
		return combined
	}
	return truncateString(state, width)
}

func (m Model) renderAuthscreen() string {
	width := m.editorContentWidth()

	help := m.Help
	help.Width = width

	rows := []string{
		renderTitleBar(width, "Authorization", m.requestContext()),
		"",
		m.AuthInput.View(),
		"",
	}
	// The decoded-token panel is the only variable-height part of this
	// screen, so it is what gives way on a short terminal.
	rows = append(rows, trimRows(m.renderTokenDetail(width), m.authDetailRows())...)
	rows = append(rows, "", m.renderAuthStatus(width), help.View(AuthKeys))

	return m.renderEditorFrame(width, rows...)
}

// renderTokenDetail decodes the token as a JWT when it is one and shows what
// it claims - the single most useful thing this screen can tell you about a
// token you just pasted.
func (m Model) renderTokenDetail(width int) []string {
	token := m.AuthInput.Value()
	if strings.TrimSpace(token) == "" {
		return []string{EditorMetaStyle.Render("No token set — requests will be sent unauthenticated.")}
	}

	jwt, ok := decodeJWT(token)
	if !ok {
		return []string{EditorMetaStyle.Render("Opaque token (not a JWT) — sent as-is.")}
	}

	now := time.Now()
	const labelWidth = 4

	header := "JWT"
	if jwt.Algorithm != "" {
		header += " · " + jwt.Algorithm
	}
	if jwt.Type != "" && !strings.EqualFold(jwt.Type, "JWT") {
		header += " · " + jwt.Type
	}

	rows := []string{EditorTitleStyle.Render(header)}

	claim := func(label, value string) {
		if value != "" {
			rows = append(rows, "  "+renderFieldRow(width-2, labelWidth, label, value, EditorValueStyle))
		}
	}
	claim("sub", jwt.Subject)
	claim("iss", jwt.Issuer)
	claim("aud", jwt.Audience)

	if !jwt.IssuedAt.IsZero() {
		claim("iat", fmt.Sprintf("%s · %s",
			relativeTime(jwt.IssuedAt, now), jwt.IssuedAt.Format("2006-01-02 15:04")))
	}

	if !jwt.ExpiresAt.IsZero() {
		style := EditorOKStyle
		note := relativeTime(jwt.ExpiresAt, now)
		if jwt.Expired(now) {
			style = EditorBadStyle
			note = "expired " + note
		} else if time.Until(jwt.ExpiresAt) < 5*time.Minute {
			style = EditorWarnStyle
		}
		rows = append(rows, "  "+renderFieldRow(width-2, labelWidth, "exp",
			fmt.Sprintf("%s · %s", note, jwt.ExpiresAt.Format("2006-01-02 15:04")), style))
	}

	return rows
}

func (m Model) renderAuthStatus(width int) string {
	if m.Status != "" {
		return EditorOKStyle.Render(truncateString(m.Status, width))
	}

	token := m.AuthInput.Value()
	if token == "" {
		return EditorMetaStyle.Render("0 characters")
	}
	return EditorMetaStyle.Render(fmt.Sprintf("%d characters", len(token)))
}
