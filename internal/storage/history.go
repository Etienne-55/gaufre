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

// MaxHistoryItems is how many requests the history keeps, newest first.
const MaxHistoryItems = 100

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

	if len(items) > MaxHistoryItems {
		items = items[:MaxHistoryItems]
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	// The history holds request bodies - login payloads, passwords and all -
	// so it is kept private to the user.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return writeFileAtomic(path, data)
}

// writeFileAtomic replaces path with data in one step: the data goes to a
// temporary file next to it (created 0600) that is then renamed over it, so
// a crash mid-write can't leave a truncated history behind.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // fails harmlessly once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
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

	if os.MkdirAll(filepath.Dir(path), 0o700) == nil && os.Rename(legacy, path) == nil {
		return path
	}
	return legacy
}
