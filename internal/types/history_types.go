package types

import (
	"fmt"
	"time"
)

type HistoryItem struct {
	Method    string
	URL       string
	Response  int
	Timestamp time.Time
	Payload   string
}

// FilterValue matches Title character for character: the history list
// highlights filter matches by their position in the title, so a different
// order (it used to be URL then method) underlined the wrong characters.
func (h HistoryItem) FilterValue() string { return h.Title() }
func (h HistoryItem) Title() string       { return h.Method + " " + h.URL }
func (h HistoryItem) Description() string {
	if when := h.When(); when != "" {
		return fmt.Sprintf("%s · %s", h.Status(), when)
	}
	return h.Status()
}

// Status is the response's status code, or a dash when none was recorded.
func (h HistoryItem) Status() string {
	if h.Response > 0 {
		return fmt.Sprintf("%d", h.Response)
	}
	return "—"
}

// When is the time the request was sent, empty when none was recorded.
func (h HistoryItem) When() string {
	if h.Timestamp.IsZero() {
		return ""
	}
	return h.Timestamp.Format("Jan 2 15:04")
}
