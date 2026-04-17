package db

import (
	"os"
	"testing"
	"time"
)

func TestDBInit(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	// Verify database is accessible
	count, err := db.GetMessageCount()
	if err != nil {
		t.Fatalf("failed to get message count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 messages, got %d", count)
	}
}

func TestInsertAndSearchMessage(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	msg := &Message{
		ID:        "test-msg-001",
		ChatID:    "1234567890@s.whatsapp.net",
		SenderID:  "0987654321@s.whatsapp.net",
		Content:   "Hello, World!",
		Timestamp: time.Now(),
		IsGroup:   false,
		HasMedia:  false,
	}

	if err := db.InsertMessage(msg); err != nil {
		t.Fatalf("failed to insert message: %v", err)
	}

	// Search for the message
	results, err := db.SearchMessages("Hello")
	if err != nil {
		t.Fatalf("failed to search messages: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Content != "Hello, World!" {
		t.Errorf("expected content 'Hello, World!', got '%s'", results[0].Content)
	}

	if results[0].ID != "test-msg-001" {
		t.Errorf("expected ID 'test-msg-001', got '%s'", results[0].ID)
	}
}

func TestInsertChat(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	chat := &Chat{
		JID:         "1234567890@s.whatsapp.net",
		Name:        "Test Chat",
		IsGroup:     false,
		LastMessage: time.Now(),
	}

	if err := db.InsertChat(chat); err != nil {
		t.Fatalf("failed to insert chat: %v", err)
	}

	chats, err := db.GetChats()
	if err != nil {
		t.Fatalf("failed to get chats: %v", err)
	}

	if len(chats) != 1 {
		t.Fatalf("expected 1 chat, got %d", len(chats))
	}

	if chats[0].Name != "Test Chat" {
		t.Errorf("expected name 'Test Chat', got '%s'", chats[0].Name)
	}

	if chats[0].JID != "1234567890@s.whatsapp.net" {
		t.Errorf("expected JID '1234567890@s.whatsapp.net', got '%s'", chats[0].JID)
	}
}

func TestInsertContact(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	contact := &Contact{
		JID:    "1234567890@s.whatsapp.net",
		Name:   "John Doe",
		Number: "+1234567890",
	}

	if err := db.InsertContact(contact); err != nil {
		t.Fatalf("failed to insert contact: %v", err)
	}

	contacts, err := db.GetContacts()
	if err != nil {
		t.Fatalf("failed to get contacts: %v", err)
	}

	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}

	if contacts[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", contacts[0].Name)
	}
}

func TestGetMessagesByChatID(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	chatID := "1234567890@s.whatsapp.net"

	// Insert multiple messages
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        "msg-" + string(rune('0'+i)),
			ChatID:    chatID,
			SenderID:  "0987654321@s.whatsapp.net",
			Content:   "Message content",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := db.InsertMessage(msg); err != nil {
			t.Fatalf("failed to insert message: %v", err)
		}
	}

	messages, err := db.GetMessagesByChatID(chatID, 10)
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}

	if len(messages) != 5 {
		t.Errorf("expected 5 messages, got %d", len(messages))
	}
}

func TestDeleteMessage(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	msg := &Message{
		ID:        "delete-me",
		ChatID:    "1234567890@s.whatsapp.net",
		SenderID:  "0987654321@s.whatsapp.net",
		Content:   "To be deleted",
		Timestamp: time.Now(),
	}

	if err := db.InsertMessage(msg); err != nil {
		t.Fatalf("failed to insert message: %v", err)
	}

	if err := db.DeleteMessage("delete-me"); err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	count, err := db.GetMessageCount()
	if err != nil {
		t.Fatalf("failed to get message count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 messages after delete, got %d", count)
	}
}

func TestFTS5Search(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	messages := []*Message{
		{ID: "m1", ChatID: "c1", Content: "Hello world", Timestamp: time.Now()},
		{ID: "m2", ChatID: "c1", Content: "Goodbye world", Timestamp: time.Now()},
		{ID: "m3", ChatID: "c1", Content: "Hello there", Timestamp: time.Now()},
		{ID: "m4", ChatID: "c1", Content: "Whats up?", Timestamp: time.Now()},
	}

	for _, m := range messages {
		if err := db.InsertMessage(m); err != nil {
			t.Fatalf("failed to insert message: %v", err)
		}
	}

	// Search for "Hello"
	results, err := db.SearchMessages("Hello")
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Hello', got %d", len(results))
	}

	// Search for "world"
	results, err = db.SearchMessages("world")
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for 'world', got %d", len(results))
	}
}

func TestGetChatByJID(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	chat := &Chat{
		JID:         "1234567890@s.whatsapp.net",
		Name:        "Test Chat",
		IsGroup:     true,
		LastMessage: time.Now(),
	}

	if err := db.InsertChat(chat); err != nil {
		t.Fatalf("failed to insert chat: %v", err)
	}

	retrieved, err := db.GetChatByJID("1234567890@s.whatsapp.net")
	if err != nil {
		t.Fatalf("failed to get chat by JID: %v", err)
	}

	if retrieved.Name != "Test Chat" {
		t.Errorf("expected 'Test Chat', got '%s'", retrieved.Name)
	}

	if !retrieved.IsGroup {
		t.Error("expected IsGroup to be true")
	}
}

func TestMessageWithMedia(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	msg := &Message{
		ID:        "media-msg",
		ChatID:    "1234567890@s.whatsapp.net",
		SenderID:  "0987654321@s.whatsapp.net",
		Content:   "Check out this image!",
		Timestamp: time.Now(),
		IsGroup:   false,
		HasMedia:  true,
		MediaType: "image",
		ReplyTo:   "",
	}

	if err := db.InsertMessage(msg); err != nil {
		t.Fatalf("failed to insert message: %v", err)
	}

	results, err := db.SearchMessages("image")
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].HasMedia {
		t.Error("expected HasMedia to be true")
	}

	if results[0].MediaType != "image" {
		t.Errorf("expected MediaType 'image', got '%s'", results[0].MediaType)
	}
}

func TestMultipleChats(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "wacli_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db := New(tmpfile.Name())
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	chats := []*Chat{
		{JID: "c1@s.whatsapp.net", Name: "Chat One", LastMessage: time.Now().Add(-1 * time.Hour)},
		{JID: "c2@s.whatsapp.net", Name: "Chat Two", LastMessage: time.Now()},
		{JID: "c3@s.whatsapp.net", Name: "Chat Three", LastMessage: time.Now().Add(-2 * time.Hour)},
	}

	for _, c := range chats {
		if err := db.InsertChat(c); err != nil {
			t.Fatalf("failed to insert chat: %v", err)
		}
	}

	retrieved, err := db.GetChats()
	if err != nil {
		t.Fatalf("failed to get chats: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 chats, got %d", len(retrieved))
	}

	// Verify order: most recent first
	if retrieved[0].Name != "Chat Two" {
		t.Errorf("expected first chat to be 'Chat Two', got '%s'", retrieved[0].Name)
	}
}
