package storage

import (
	"testing"
	"time"

	"gaufre/internal/types"
)

func TestLoadHistory_MissingFileReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir()) // Windows equivalent used by os.UserHomeDir

	items, err := LoadHistory()
	if err != nil {
		t.Fatalf("unexpected error for a missing history file: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %v, want empty", items)
	}
}

func TestSaveAndLoadHistory_Roundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())

	want := []types.HistoryItem{
		{Method: "GET", URL: "http://example.com", Response: 200, Timestamp: time.Now().Truncate(time.Second)},
		{Method: "POST", URL: "http://example.com/x", Payload: `{"a":1}`, Response: 201},
	}

	if err := SaveHistory(want); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	got, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Method != want[i].Method || got[i].URL != want[i].URL || got[i].Payload != want[i].Payload || got[i].Response != want[i].Response {
			t.Errorf("item %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSaveHistory_TruncatesToMax(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())

	items := make([]types.HistoryItem, maxHistoryItems+10)
	for i := range items {
		items[i] = types.HistoryItem{Method: "GET", URL: "http://example.com"}
	}

	if err := SaveHistory(items); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	got, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(got) != maxHistoryItems {
		t.Fatalf("len(got) = %d, want %d", len(got), maxHistoryItems)
	}
}
