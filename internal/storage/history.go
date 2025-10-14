package storage

import (
	"os"
	"encoding/json"
	"path/filepath"
	"gaufre/internal/types"
)


const historyFileName = ".url_history.json"

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

	data, err := json.MarshalIndent(items, "", "  ")
	if err  != nil {
		return  err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadHistory() ([]types.HistoryItem, error) {
	path, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsExist(err) {
		return []types.HistoryItem{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []types.HistoryItem
	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

