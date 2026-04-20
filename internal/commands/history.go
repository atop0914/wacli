package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "History operations",
	Long:  `Commands for managing message history`,
}

var historyBackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Request message history",
	Long:  `Request older messages from a chat or all chats`,
	RunE:  runHistoryBackfill,
}

// historyFlags holds the flags for history commands
var historyFlags = struct {
	chat       string
	count      int
	jsonOutput bool
}{}

func init() {
	historyBackfillCmd.Flags().StringVarP(&historyFlags.chat, "chat", "c", "", "Chat JID to backfill (empty for all chats)")
	historyBackfillCmd.Flags().IntVarP(&historyFlags.count, "count", "n", 100, "Number of messages to request")
	historyBackfillCmd.Flags().BoolVar(&historyFlags.jsonOutput, "json", false, "Output in JSON format")

	historyCmd.AddCommand(historyBackfillCmd)
}

// HistoryBackfillResult represents the result of a backfill request
type HistoryBackfillResult struct {
	Success    bool      `json:"success"`
	ChatJID     string    `json:"chat_jid,omitempty"`
	MessageCount int      `json:"message_count"`
	StartTime   time.Time `json:"start_time"`
	Error       string    `json:"error,omitempty"`
}

// HistoryRequest represents a history request for display
type HistoryRequest struct {
	JID        string    `json:"jid"`
	Count      int       `json:"count"`
	Timestamp  time.Time `json:"timestamp"`
	Completed  bool      `json:"completed"`
	MessagesReceived int `json:"messages_received"`
}

func runHistoryBackfill(cmd *cobra.Command, args []string) error {
	waClient := GetClient()
	if waClient == nil {
		return fmt.Errorf("client not initialized. Run 'wacli auth' first")
	}

	if !waClient.IsConnected() {
		return fmt.Errorf("client not connected. Run 'wacli auth' first")
	}

	ctx := context.Background()
	startTime := time.Now()

	var err error
	if historyFlags.chat != "" {
		// Backfill specific chat
		err = waClient.RequestHistory(ctx, historyFlags.chat, historyFlags.count)
	} else {
		// Backfill all chats - get groups first
		groups, groupErr := waClient.GetGroups(ctx)
		if groupErr != nil {
			err = groupErr
		} else {
			for _, g := range groups {
				if reqErr := waClient.RequestHistory(ctx, g.JID.String(), historyFlags.count); reqErr != nil {
					fmt.Fprintf(os.Stderr, "Failed to backfill %s: %s\n", g.JID.String(), reqErr.Error())
				}
			}
		}
	}

	result := HistoryBackfillResult{
		Success:     err == nil,
		ChatJID:     historyFlags.chat,
		MessageCount: historyFlags.count,
		StartTime:   startTime,
	}
	if err != nil {
		result.Error = err.Error()
	}

	if historyFlags.jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		if result.Success {
			if historyFlags.chat != "" {
				fmt.Printf("✓ History backfill requested for %s\n", historyFlags.chat)
				fmt.Printf("  Requesting %d messages\n", historyFlags.count)
			} else {
				fmt.Printf("✓ History backfill requested for all chats\n")
				fmt.Printf("  Requesting %d messages per chat\n", historyFlags.count)
			}
		} else {
			fmt.Fprintf(os.Stderr, "✗ Backfill failed: %s\n", result.Error)
		}
	}

	return err
}