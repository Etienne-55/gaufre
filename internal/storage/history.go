package storage

import (
	"encoding/json"
	"errors"
	"gaufre/internal/types"
	"io/fs"
	"os"
	"path/filepath"
)

const historyFileName = "history.json"
const maxHistoryItems = 100

// legacyHistoryFileName is where the history used to live, loose in the
// home directory. LoadHistory moves it into the config directory.
const legacyHistoryFileName = ".url_history.json"

// ConfigDir is Gaufre's folder next to the other tools' configs:
// $XDG_CONFIG_HOME/gaufre, or ~/.config/gaufre. That holds on macOS too,
// rather than Go's os.UserConfigDir (~/Library/Application Support), where
// nobody looks for a terminal tool's files.
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "gaufre"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gaufre"), nil
}

func GetHistoryPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, historyFileName), nil
}

func SaveHistory(items []types.HistoryItem) error {
	path, err := GetHistoryPath()
	if err != nil {
		return err
	}

	if len(items) > maxHistoryItems {
		items = items[:maxHistoryItems]
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadHistory() ([]types.HistoryItem, error) {
	path, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(historyToLoad(path))
	if err != nil {
		if os.IsNotExist(err) {
			return []types.HistoryItem{}, nil
		}
		return nil, err
	}

	var items []types.HistoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// historyToLoad is the file to read the history from. The first time there
// is none at path, a history left at its old place in the home directory is
// moved there; if it can't be moved it is read where it is, and the next
// save writes it to path anyway.
func historyToLoad(path string) string {
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	legacy := filepath.Join(home, legacyHistoryFileName)
	if _, err := os.Stat(legacy); err != nil {
		return path
	}

	if os.MkdirAll(filepath.Dir(path), 0o755) == nil && os.Rename(legacy, path) == nil {
		return path
	}
	return legacy
}
