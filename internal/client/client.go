package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/atop0914/wacli/internal/db"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// MessageHandler is called when a message is received
type MessageHandler func(msg *db.Message)

// WhatsAppClient wraps the whatsmeow client with additional functionality
type WhatsAppClient struct {
	client      *whatsmeow.Client
	db          *db.DB
	mu          sync.RWMutex
	isConnected bool
	onMessage   MessageHandler
}

// New creates a new WhatsApp client wrapper
func New(database *db.DB) *WhatsAppClient {
	return &WhatsAppClient{
		db:          database,
		isConnected: false,
	}
}

// AttachClient attaches a whatsmeow client to this wrapper
func (c *WhatsAppClient) AttachClient(client *whatsmeow.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.client = client
}

// SetMessageHandler sets the callback for incoming messages
func (c *WhatsAppClient) SetMessageHandler(handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = handler
}

// Connect establishes connection to WhatsApp servers
func (c *WhatsAppClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return fmt.Errorf("client not initialized")
	}

	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.isConnected = true
	return nil
}

// Disconnect closes the WhatsApp connection
func (c *WhatsAppClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil
	}

	c.client.Disconnect()
	c.isConnected = false
	return nil
}

// IsConnected returns whether the client is connected
func (c *WhatsAppClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// SendTextMessage sends a text message to a recipient
func (c *WhatsAppClient) SendTextMessage(ctx context.Context, to string, message string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil || !c.isConnected {
		return fmt.Errorf("client not connected")
	}

	jid, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	_, err = c.client.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: &message,
	})

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// GetGroups retrieves all joined groups
func (c *WhatsAppClient) GetGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil || !c.isConnected {
		return nil, fmt.Errorf("client not connected")
	}

	return c.client.GetJoinedGroups(ctx)
}

// GetGroupByJID retrieves a specific group by JID
func (c *WhatsAppClient) GetGroupByJID(ctx context.Context, jid string) (*types.GroupInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil || !c.isConnected {
		return nil, fmt.Errorf("client not connected")
	}

	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	return c.client.GetGroupInfo(ctx, parsedJID)
}

// HandleIncomingMessage processes an incoming message event
func (c *WhatsAppClient) HandleIncomingMessage(evt *events.Message) {
	c.mu.RLock()
	handler := c.onMessage
	c.mu.RUnlock()

	msg := &db.Message{
		ID:       string(evt.Info.ID),
		ChatID:   evt.Info.Chat.String(),
		SenderID: evt.Info.Sender.String(),
		Timestamp: evt.Info.Timestamp,
		IsGroup:  evt.Info.Chat.Server == types.GroupServer,
	}

	// Extract text content
	if evt.Message.GetConversation() != "" {
		msg.Content = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		msg.Content = evt.Message.GetExtendedTextMessage().GetText()
	}

	// Check for media
	if evt.Message.GetImageMessage() != nil {
		msg.HasMedia = true
		msg.MediaType = "image"
	} else if evt.Message.GetVideoMessage() != nil {
		msg.HasMedia = true
		msg.MediaType = "video"
	} else if evt.Message.GetAudioMessage() != nil {
		msg.HasMedia = true
		msg.MediaType = "audio"
	} else if evt.Message.GetDocumentMessage() != nil {
		msg.HasMedia = true
		msg.MediaType = "document"
	}

	// Write to database
	if c.db != nil {
		if err := c.db.InsertMessage(msg); err != nil {
			fmt.Printf("Failed to insert message: %v\n", err)
		}
	}

	if handler != nil {
		handler(msg)
	}
}

// GetClient returns the underlying whatsmeow client
func (c *WhatsAppClient) GetClient() *whatsmeow.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// GetMyJID returns the JID of the logged-in device
func (c *WhatsAppClient) GetMyJID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.client == nil || c.client.Store == nil {
		return ""
	}
	return c.client.Store.ID.String()
}

// MarshalJSON implements json.Marshaler for WhatsAppClient
func (c *WhatsAppClient) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	type ClientJSON struct {
		IsConnected bool   `json:"is_connected"`
		MyJID       string `json:"my_jid"`
	}

	return json.Marshal(ClientJSON{
		IsConnected: c.isConnected,
		MyJID:       c.GetMyJID(),
	})
}