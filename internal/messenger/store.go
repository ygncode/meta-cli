package messenger

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS messages (
	id           TEXT PRIMARY KEY,
	psid         TEXT NOT NULL,
	page_id      TEXT NOT NULL,
	text         TEXT NOT NULL,
	direction    TEXT NOT NULL DEFAULT 'in',
	auto_replied INTEGER NOT NULL DEFAULT 0,
	received_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_psid ON messages(psid);
CREATE INDEX IF NOT EXISTS idx_messages_page ON messages(page_id);
CREATE INDEX IF NOT EXISTS idx_messages_ts   ON messages(received_at DESC);
`

type Store struct {
	db *sql.DB
}

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".meta-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "messages.db"), nil
}

func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveMessage(m *Message) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO messages (id, psid, page_id, text, direction, auto_replied, received_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.PSID, m.PageID, m.Text, m.Direction, m.AutoReplied, m.ReceivedAt.Unix(),
	)
	return err
}

func (s *Store) ListMessages(pageID string, limit int) ([]Message, error) {
	rows, err := s.db.Query(
		`SELECT id, psid, page_id, text, direction, auto_replied, received_at FROM messages WHERE page_id = ? ORDER BY received_at DESC LIMIT ?`,
		pageID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var ts int64
		var autoReplied int
		if err := rows.Scan(&m.ID, &m.PSID, &m.PageID, &m.Text, &m.Direction, &autoReplied, &ts); err != nil {
			return nil, err
		}
		m.ReceivedAt = time.Unix(ts, 0)
		m.AutoReplied = autoReplied != 0
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *Store) MessageExists(id string) bool {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM messages WHERE id = ?`, id).Scan(&count); err != nil {
		log.Printf("MessageExists query error: %v", err)
	}
	return count > 0
}

func (s *Store) MarkAutoReplied(msgID string) error {
	_, err := s.db.Exec(`UPDATE messages SET auto_replied = 1 WHERE id = ?`, msgID)
	return err
}
