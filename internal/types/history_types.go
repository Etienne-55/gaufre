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

func (h HistoryItem) FilterValue() string { return h.URL + " " + h.Method }
func (h HistoryItem) Title() string       { return h.Method + " " + h.URL }
func (h HistoryItem) Description() string {
	status := "—"
	if h.Response > 0 {
		status = fmt.Sprintf("%d", h.Response)
	}

	if h.Timestamp.IsZero() {
		return status
	}

	return fmt.Sprintf("%s · %s", status, h.Timestamp.Format("Jan 2 15:04"))
}
