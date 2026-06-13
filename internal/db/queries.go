package db

import (
	"database/sql"
	"time"
)

type HistoryItem struct {
	ID        int
	Command   string
	Args      string
	Output    string
	Success   bool
	Timestamp time.Time
}

func (db *DB) AddHistory(command, args, output string, success bool) error {
	_, err := db.Exec(
		"INSERT INTO history (command, args, output, success) VALUES (?, ?, ?, ?)",
		command, args, output, success,
	)
	return err
}

func (db *DB) GetHistory(limit int) ([]HistoryItem, error) {
	rows, err := db.Query(
		"SELECT id, command, args, output, success, timestamp FROM history ORDER BY timestamp DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var item HistoryItem
		if err := rows.Scan(&item.ID, &item.Command, &item.Args, &item.Output, &item.Success, &item.Timestamp); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *DB) SetSetting(key, value string) error {
	_, err := db.Exec(
		"INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

func (db *DB) GetAIConfig(provider string) (string, error) {
	var config string
	err := db.QueryRow("SELECT config FROM ai_config WHERE provider = ?", provider).Scan(&config)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return config, err
}

func (db *DB) SetAIConfig(provider, config string) error {
	_, err := db.Exec(
		"INSERT OR REPLACE INTO ai_config (provider, config) VALUES (?, ?)",
		provider, config,
	)
	return err
}

type ChatMessage struct {
	Role    string
	Content string
}

func (db *DB) AddChatMessage(provider, role, content string) error {
	_, err := db.Exec(
		"INSERT INTO chat_history (provider, role, content) VALUES (?, ?, ?)",
		provider, role, content,
	)
	return err
}

func (db *DB) GetChatHistory(provider string, limit int) ([]ChatMessage, error) {
	rows, err := db.Query(
		"SELECT role, content FROM chat_history WHERE provider = ? ORDER BY created_at DESC LIMIT ?",
		provider, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
