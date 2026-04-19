package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/atop0914/wacli/internal/auth"
	"github.com/atop0914/wacli/internal/client"
	"github.com/atop0914/wacli/internal/commands"
	"github.com/atop0914/wacli/internal/db"
	"github.com/atop0914/wacli/internal/store"
	"github.com/spf13/cobra"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var (
	storeDir   string
	deviceName string
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "wacli",
	Short: "WhatsApp CLI - A terminal client for WhatsApp",
	Long: `wacli is a command-line interface for WhatsApp, built on top of whatsmeow.
It supports message sync, search, sending, and group management.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initServices()
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with WhatsApp",
	Long:  `Display QR code and authenticate with WhatsApp Web`,
	RunE:  runAuth,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&storeDir, "store", "", "Storage directory (default: ~/.wacli)")
	rootCmd.PersistentFlags().StringVar(&deviceName, "device-name", "", "Device label")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// Add commands
	rootCmd.AddCommand(authCmd)

	// Register commands package commands
	commands.RegisterCommands(rootCmd)
}

func main() {
	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		commands.StopSync()
		os.Exit(0)
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var (
	authManager *auth.AuthManager
	waClient    *client.WhatsAppClient
	database    *db.DB
)

func initServices() error {
	// Get store directory
	storePath := store.GetStoreDir(storeDir)
	if err := store.EnsureStoreDir(storePath); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Initialize database
	dbPath := filepath.Join(storePath, "wacli.db")
	database = db.New(dbPath)
	if err := database.Init(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Set database in commands package
	commands.SetDB(database)

	// Create logger
	logger := waLog.Stdout("wacli", "DEBUG", true)

	// Initialize WhatsApp client
	waClient = client.New(database)
	commands.SetClient(waClient)

	// Create event handler
	eventHandler := auth.NewEventHandler(
		func(qr []string) {
			for i, code := range qr {
				fmt.Printf("QR Code %d: %s\n", i+1, code)
			}
		},
		func(deviceJID string) {
			fmt.Printf("Logged in as: %s\n", deviceJID)
		},
	)

	// Create auth manager
	am := auth.NewAuthManager(nil, logger, eventHandler)
	if err := am.NewClient(); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	authManager = am
	waClient.AttachClient(am.GetClient())

	// Set up message handler to write to DB
	waClient.SetMessageHandler(func(msg *db.Message) {
		if err := database.InsertMessage(msg); err != nil {
			logger.Errorf("Failed to insert message: %v", err)
		}
	})

	return nil
}

func runAuth(cmd *cobra.Command, args []string) error {
	if authManager == nil {
		return fmt.Errorf("client not initialized")
	}

	ctx := context.Background()
	if err := authManager.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Keep running to maintain connection
	select {}
}
