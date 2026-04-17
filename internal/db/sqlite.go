package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// Message represents a WhatsApp message
type Message struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsGroup   bool      `json:"is_group"`
	HasMedia  bool      `json:"has_media"`
	MediaType string    `json:"media_type,omitempty"`
	ReplyTo   string    `json:"reply_to,omitempty"`
}

// Chat represents a WhatsApp chat
type Chat struct {
	JID         string    `json:"jid"`
	Name        string    `json:"name"`
	IsGroup     bool      `json:"is_group"`
	LastMessage time.Time `json:"last_message"`
}

// Contact represents a WhatsApp contact
type Contact struct {
	JID    string `json:"jid"`
	Name   string `json:"name"`
	Number string `json:"number"`
}

// New creates a new DB instance
func New(path string) *DB {
	return &DB{path: path}
}

// Init initializes the database connection and runs migrations
func (db *DB) Init() error {
	conn, err := sql.Open("sqlite3", db.path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	db.conn = conn

	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := db.migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	migrations := []string{
		// Chats table
		`CREATE TABLE IF NOT EXISTS chats (
			jid TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			is_group INTEGER NOT NULL DEFAULT 0,
			last_message_at INTEGER
		)`,
		// Contacts table
		`CREATE TABLE IF NOT EXISTS contacts (
			jid TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			number TEXT NOT NULL DEFAULT ''
		)`,
		// Messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			chat_id TEXT NOT NULL,
			sender_id TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			timestamp INTEGER NOT NULL,
			is_group INTEGER NOT NULL DEFAULT 0,
			has_media INTEGER NOT NULL DEFAULT 0,
			media_type TEXT NOT NULL DEFAULT '',
			reply_to TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (chat_id) REFERENCES chats(jid)
		)`,
		// Index on chat_id for faster message lookups
		`CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id)`,
		// Index on timestamp for chronological queries
		`CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC)`,
		// FTS5 virtual table for full-text search
		`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
			content,
			content='messages',
			content_rowid='rowid'
		)`,
		// Triggers to keep FTS5 in sync with messages table
		`CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, content) VALUES (NEW.rowid, NEW.content);
		 END`,
		`CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', OLD.rowid, OLD.content);
		 END`,
		`CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', OLD.rowid, OLD.content);
			INSERT INTO messages_fts(rowid, content) VALUES (NEW.rowid, NEW.content);
		 END`,
		// Schema version for future migrations
		`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`,
	}

	for i, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Set initial schema version
	var version int
	row := db.conn.QueryRow("SELECT version FROM schema_version LIMIT 1")
	if err := row.Scan(&version); err == sql.ErrNoRows {
		if _, err := db.conn.Exec("INSERT INTO schema_version (version) VALUES (?)", len(migrations)); err != nil {
			return fmt.Errorf("failed to set schema version: %w", err)
		}
	}

	return nil
}

// InsertMessage inserts a new message into the database
func (db *DB) InsertMessage(msg *Message) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO messages (id, chat_id, sender_id, content, timestamp, is_group, has_media, media_type, reply_to)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID,
		msg.ChatID,
		msg.SenderID,
		msg.Content,
		msg.Timestamp.UnixMilli(),
		boolToInt(msg.IsGroup),
		boolToInt(msg.HasMedia),
		msg.MediaType,
		msg.ReplyTo,
	)
	return err
}

// SearchMessages searches messages using FTS5
func (db *DB) SearchMessages(query string) ([]*Message, error) {
	rows, err := db.conn.Query(`
		SELECT m.id, m.chat_id, m.sender_id, m.content, m.timestamp, m.is_group, m.has_media, m.media_type, m.reply_to
		FROM messages m
		INNER JOIN messages_fts fts ON m.rowid = fts.rowid
		WHERE messages_fts MATCH ?
		ORDER BY m.timestamp DESC
		LIMIT 100`,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

// GetMessagesByChatID retrieves all messages for a specific chat
func (db *DB) GetMessagesByChatID(chatID string, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.conn.Query(`
		SELECT id, chat_id, sender_id, content, timestamp, is_group, has_media, media_type, reply_to
		FROM messages
		WHERE chat_id = ?
		ORDER BY timestamp DESC
		LIMIT ?`,
		chatID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

// InsertChat inserts or updates a chat
func (db *DB) InsertChat(chat *Chat) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO chats (jid, name, is_group, last_message_at)
		VALUES (?, ?, ?, ?)`,
		chat.JID,
		chat.Name,
		boolToInt(chat.IsGroup),
		chat.LastMessage.UnixMilli(),
	)
	return err
}

// GetChats retrieves all chats
func (db *DB) GetChats() ([]*Chat, error) {
	rows, err := db.conn.Query(`
		SELECT jid, name, is_group, last_message_at
		FROM chats
		ORDER BY last_message_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var chat Chat
		var lastMsgAt int64
		var isGroup int
		if err := rows.Scan(&chat.JID, &chat.Name, &isGroup, &lastMsgAt); err != nil {
			return nil, err
		}
		chat.IsGroup = intToBool(isGroup)
		chat.LastMessage = time.UnixMilli(lastMsgAt)
		chats = append(chats, &chat)
	}
	return chats, rows.Err()
}

// GetChatByJID retrieves a single chat by JID
func (db *DB) GetChatByJID(jid string) (*Chat, error) {
	var chat Chat
	var lastMsgAt int64
	var isGroup int
	err := db.conn.QueryRow(`
		SELECT jid, name, is_group, last_message_at
		FROM chats
		WHERE jid = ?`, jid,
	).Scan(&chat.JID, &chat.Name, &isGroup, &lastMsgAt)
	if err != nil {
		return nil, err
	}
	chat.IsGroup = intToBool(isGroup)
	chat.LastMessage = time.UnixMilli(lastMsgAt)
	return &chat, nil
}

// InsertContact inserts or updates a contact
func (db *DB) InsertContact(contact *Contact) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO contacts (jid, name, number)
		VALUES (?, ?, ?)`,
		contact.JID,
		contact.Name,
		contact.Number,
	)
	return err
}

// GetContacts retrieves all contacts
func (db *DB) GetContacts() ([]*Contact, error) {
	rows, err := db.conn.Query(`SELECT jid, name, number FROM contacts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.JID, &c.Name, &c.Number); err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, rows.Err()
}

// DeleteMessage deletes a message by ID
func (db *DB) DeleteMessage(id string) error {
	_, err := db.conn.Exec(`DELETE FROM messages WHERE id = ?`, id)
	return err
}

// GetMessageCount returns the total number of messages
func (db *DB) GetMessageCount() (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count)
	return count, err
}

// Helper functions

func scanMessages(rows *sql.Rows) ([]*Message, error) {
	var messages []*Message
	for rows.Next() {
		var msg Message
		var timestamp int64
		var isGroup, hasMedia int
		if err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.SenderID,
			&msg.Content,
			&timestamp,
			&isGroup,
			&hasMedia,
			&msg.MediaType,
			&msg.ReplyTo,
		); err != nil {
			return nil, err
		}
		msg.IsGroup = intToBool(isGroup)
		msg.HasMedia = intToBool(hasMedia)
		msg.Timestamp = time.UnixMilli(timestamp)
		messages = append(messages, &msg)
	}
	return messages, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
