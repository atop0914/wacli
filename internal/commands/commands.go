package commands

import (
	"context"
	"fmt"
	"sync"

	"github.com/atop0914/wacli/internal/db"
	"github.com/atop0914/wacli/internal/client"
	"github.com/spf13/cobra"
	"go.mau.fi/whatsmeow/types"
)

var (
	waClient *WhatsAppClientWrapper
	database *db.DB
	clientOnce sync.Once
)

// WhatsAppClientWrapper wraps the WhatsApp client with simplified interface
type WhatsAppClientWrapper struct {
	client *client.WhatsAppClient
}

// IsConnected returns whether the underlying client is connected
func (w *WhatsAppClientWrapper) IsConnected() bool {
	if w == nil || w.client == nil {
		return false
	}
	return w.client.IsConnected()
}

// SendTextMessage sends a text message
func (w *WhatsAppClientWrapper) SendTextMessage(ctx context.Context, to string, message string) error {
	if w == nil || w.client == nil {
		return fmt.Errorf("client not initialized")
	}
	return w.client.SendTextMessage(ctx, to, message)
}

// SendFile sends a file
func (w *WhatsAppClientWrapper) SendFile(ctx context.Context, to string, filePath string, caption string) error {
	if w == nil || w.client == nil {
		return fmt.Errorf("client not initialized")
	}
	return w.client.SendFile(ctx, to, filePath, caption)
}

// SetMessageHandler sets the message handler
func (w *WhatsAppClientWrapper) SetMessageHandler(handler func(msg *db.Message)) {
	if w != nil && w.client != nil {
		w.client.SetMessageHandler(handler)
	}
}

// GetGroups retrieves groups
func (w *WhatsAppClientWrapper) GetGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if w == nil || w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	return w.client.GetGroups(ctx)
}

// GetClient returns the global WhatsApp client instance
func GetClient() *WhatsAppClientWrapper {
	return waClient
}

// SetClient sets the global WhatsApp client instance
func SetClient(c *client.WhatsAppClient) {
	waClient = &WhatsAppClientWrapper{client: c}
}

// GetDB returns the global database instance
func GetDB() *db.DB {
	return database
}

// SetDB sets the global database instance
func SetDB(d *db.DB) {
	database = d
}

// GetCommand returns the root commands command
func GetCommand() *cobra.Command {
	return sendCmd
}

// RegisterCommands registers all commands with the root command
func RegisterCommands(root *cobra.Command) {
	root.AddCommand(sendCmd)
	root.AddCommand(messagesCmd)
	root.AddCommand(syncCmd)
}
