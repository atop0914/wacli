package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// ErrNotAuthenticated is returned when trying to use a non-authenticated client
var ErrNotAuthenticated = errors.New("not authenticated")

// ErrAlreadyLoggedIn is returned when trying to login while already logged in
var ErrAlreadyLoggedIn = errors.New("already logged in")

// QRCallback is called when QR code is generated (codes are string arrays)
type QRCallback func(qr []string)

// LoginCallback is called when login is successful
type LoginCallback func(deviceJID string)

// MessageCallback is called when a message is received
type MessageCallback func(msg *ReceivedMessage)

// EventHandler handles WhatsApp events
type EventHandler struct {
	mu            sync.RWMutex
	onQR          QRCallback
	onLogin       LoginCallback
	onMessage     MessageCallback
	onConnected   func()
	onDisconnected func()
}

// NewEventHandler creates a new event handler
func NewEventHandler(qrCb QRCallback, loginCb LoginCallback) *EventHandler {
	return &EventHandler{
		onQR:    qrCb,
		onLogin: loginCb,
	}
}

// ReceivedMessage represents an incoming message
type ReceivedMessage struct {
	ID        string
	ChatID    string
	SenderID  string
	Content   string
	IsGroup   bool
	HasMedia  bool
	MediaType string
}

// AuthManager handles WhatsApp authentication
type AuthManager struct {
	client   *whatsmeow.Client
	device   *store.Device
	isLogged bool
	mu       sync.RWMutex
	handler  *EventHandler
	logger   waLog.Logger
}

// NewAuthManager creates a new auth manager with a device store
func NewAuthManager(device *store.Device, logger waLog.Logger, handler *EventHandler) *AuthManager {
	return &AuthManager{
		device:  device,
		logger:  logger,
		handler: handler,
	}
}

// NewClient creates a new WhatsApp client
func (am *AuthManager) NewClient() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.client != nil {
		return ErrAlreadyLoggedIn
	}

	am.client = whatsmeow.NewClient(am.device, am.logger)
	return nil
}

// GetClient returns the WhatsApp client
func (am *AuthManager) GetClient() *whatsmeow.Client {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.client
}

// IsAuthenticated checks if the client is authenticated
func (am *AuthManager) IsAuthenticated() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.isLogged && am.client != nil && am.client.IsConnected()
}

// handleEvent dispatches WhatsApp events
func (am *AuthManager) handleEvent(evt interface{}) {
	if am.handler == nil {
		return
	}

	switch v := evt.(type) {
	case *events.QR:
		if am.handler.onQR != nil {
			am.handler.onQR(v.Codes)
		}
	case *events.Message:
		if am.handler.onMessage != nil {
			msg := &ReceivedMessage{
				ID:       string(v.Info.ID),
				ChatID:   v.Info.Chat.String(),
				SenderID: v.Info.Sender.String(),
				IsGroup:  v.Info.Chat.Server == types.GroupServer,
			}
			// Extract text content
			if v.Message.GetConversation() != "" {
				msg.Content = v.Message.GetConversation()
			} else if v.Message.GetExtendedTextMessage() != nil {
				msg.Content = v.Message.GetExtendedTextMessage().GetText()
			}
			// Check for media
			if v.Message.GetImageMessage() != nil {
				msg.HasMedia = true
				msg.MediaType = "image"
			} else if v.Message.GetVideoMessage() != nil {
				msg.HasMedia = true
				msg.MediaType = "video"
			} else if v.Message.GetAudioMessage() != nil {
				msg.HasMedia = true
				msg.MediaType = "audio"
			} else if v.Message.GetDocumentMessage() != nil {
				msg.HasMedia = true
				msg.MediaType = "document"
			}
			am.handler.onMessage(msg)
		}
	case *events.Connected:
		am.mu.Lock()
		am.isLogged = true
		am.mu.Unlock()
		if am.handler.onConnected != nil {
			am.handler.onConnected()
		}
	case *events.Disconnected:
		if am.handler.onDisconnected != nil {
			am.handler.onDisconnected()
		}
	}
}

// Login initiates QR code login
func (am *AuthManager) Login(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.isLogged && am.client != nil && am.client.IsConnected() {
		return ErrAlreadyLoggedIn
	}

	if am.client == nil {
		return fmt.Errorf("client not initialized, call NewClient first")
	}

	// Connect to WhatsApp - this will trigger QR event if not logged in
	return am.client.Connect()
}

// Logout disconnects and logs out the current session
func (am *AuthManager) Logout() error {
	return am.LogoutWithContext(context.Background())
}

// LogoutWithContext disconnects and logs out the current session with context
func (am *AuthManager) LogoutWithContext(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.client == nil {
		return ErrNotAuthenticated
	}

	am.client.Logout(ctx)
	am.isLogged = false
	return nil
}

// Connect establishes connection to WhatsApp
func (am *AuthManager) Connect(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.client == nil {
		return fmt.Errorf("client not initialized")
	}

	return am.client.Connect()
}

// Disconnect closes the connection
func (am *AuthManager) Disconnect() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.client == nil {
		return nil
	}

	am.client.Disconnect()
	return nil
}

// GetDeviceJID returns the logged in device JID
func (am *AuthManager) GetDeviceJID() string {
	am.mu.RLock()
	defer am.mu.RUnlock()
	if am.client == nil {
		return ""
	}
	return am.client.Store.ID.String()
}

// SetMessageHandler sets the callback for incoming messages
func (am *AuthManager) SetMessageHandler(handler MessageCallback) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.handler != nil {
		am.handler.onMessage = handler
	}
}

// SetConnectedHandler sets the callback for connection events
func (am *AuthManager) SetConnectedHandler(handler func()) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.handler != nil {
		am.handler.onConnected = handler
	}
}

// SetDisconnectedHandler sets the callback for disconnection events
func (am *AuthManager) SetDisconnectedHandler(handler func()) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.handler != nil {
		am.handler.onDisconnected = handler
	}
}

// GenerateMockQR generates a mock QR code for testing
func GenerateMockQR() []string {
	return []string{
		"MockQR:example1",
		"MockQR:example2",
		"MockQR:example3",
		"MockQR:example4",
	}
}