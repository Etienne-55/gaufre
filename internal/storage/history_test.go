package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gaufre/internal/types"
)

// tempHome points the home directory at an empty temporary one, with no
// XDG_CONFIG_HOME, so the tests never touch the real history.
func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows equivalent used by os.UserHomeDir
	t.Setenv("XDG_CONFIG_HOME", "")
	return home
}

func TestLoadHistory_MissingFileReturnsEmpty(t *testing.T) {
	tempHome(t)

	items, err := LoadHistory()
	if err != nil {
		t.Fatalf("unexpected error for a missing history file: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %v, want empty", items)
	}
}

func TestSaveAndLoadHistory_Roundtrip(t *testing.T) {
	tempHome(t)

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
	tempHome(t)

	items := make([]types.HistoryItem, MaxHistoryItems+10)
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
	if len(got) != MaxHistoryItems {
		t.Fatalf("len(got) = %d, want %d", len(got), MaxHistoryItems)
	}
}

func TestHistoryLivesInTheConfigDir(t *testing.T) {
	home := tempHome(t)
	if err := SaveHistory([]types.HistoryItem{{Method: "GET", URL: "http://example.com"}}); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "gaufre", "history.json")); err != nil {
		t.Errorf("history not saved to ~/.config/gaufre/history.json: %v", err)
	}

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if path, _ := GetHistoryPath(); path != filepath.Join(xdg, "gaufre", "history.json") {
		t.Errorf("GetHistoryPath = %q, want it under XDG_CONFIG_HOME", path)
	}
}

// A history from before the move to the config directory is carried over,
// not lost.
func TestLoadHistory_MovesTheOldHistoryFile(t *testing.T) {
	home := tempHome(t)
	legacy := filepath.Join(home, ".url_history.json")
	if err := os.WriteFile(legacy, []byte(`[{"Method":"POST","URL":"http://old.example.com"}]`), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(items) != 1 || items[0].URL != "http://old.example.com" {
		t.Fatalf("items = %+v, want the old history", items)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Error("the old history file was left in the home directory")
	}
	path, _ := GetHistoryPath()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the old history wasn't moved to %s: %v", path, err)
	}
}

// Once there is a history in the config directory, an old file lying around
// is left alone rather than replacing it.
func TestLoadHistory_KeepsTheCurrentHistoryOverAnOldOne(t *testing.T) {
	home := tempHome(t)
	if err := SaveHistory([]types.HistoryItem{{Method: "GET", URL: "http://new.example.com"}}); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}
	legacy := filepath.Join(home, ".url_history.json")
	if err := os.WriteFile(legacy, []byte(`[{"Method":"POST","URL":"http://old.example.com"}]`), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(items) != 1 || items[0].URL != "http://new.example.com" {
		t.Errorf("items = %+v, want the config directory's history", items)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Error("the old history file was removed")
	}
}
