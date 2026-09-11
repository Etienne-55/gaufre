package storage

import (
	"encoding/json"
	"gaufre/internal/types"
	"os"
	"path/filepath"
)

const historyFileName = ".url_history.json"
const maxHistoryItems = 100

func GetHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, historyFileName), nil
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

	return os.WriteFile(path, data, 0644)
}

func LoadHistory() ([]types.HistoryItem, error) {
	path, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
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
