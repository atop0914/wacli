package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/atop0914/wacli/internal/db"
	"github.com/spf13/cobra"
)

// messagesCmd represents the messages command
var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Message operations",
	Long:  `Commands for working with messages`,
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search messages",
	Long:  `Search messages by keyword using full-text search`,
	RunE:  runSearch,
}

// searchFlags holds the flags for search command
var searchFlags = struct {
	query      string
	chat       string
	limit      int
	jsonOutput bool
}{}

func init() {
	searchCmd.Flags().StringVarP(&searchFlags.query, "query", "q", "", "Search query (required)")
	searchCmd.Flags().StringVarP(&searchFlags.chat, "chat", "c", "", "Filter by chat JID")
	searchCmd.Flags().IntVarP(&searchFlags.limit, "limit", "l", 50, "Maximum number of results")
	searchCmd.Flags().BoolVar(&searchFlags.jsonOutput, "json", false, "Output in JSON format")

	searchCmd.MarkFlagRequired("query")

	messagesCmd.AddCommand(searchCmd)
}

// SearchResult represents a formatted search result
type SearchResult struct {
	Query   string         `json:"query"`
	Count   int            `json:"count"`
	Results []*MessageInfo `json:"results"`
}

// MessageInfo represents formatted message information
type MessageInfo struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsGroup   bool      `json:"is_group"`
	HasMedia  bool      `json:"has_media"`
	MediaType string    `json:"media_type,omitempty"`
}

func runSearch(cmd *cobra.Command, args []string) error {
	database := GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	messages, err := database.SearchMessages(searchFlags.query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Apply filters
	var filtered []*db.Message
	for _, msg := range messages {
		if searchFlags.chat != "" && msg.ChatID != searchFlags.chat {
			continue
		}
		filtered = append(filtered, msg)
		if len(filtered) >= searchFlags.limit {
			break
		}
	}

	// Format results
	results := make([]*MessageInfo, len(filtered))
	for i, msg := range filtered {
		results[i] = &MessageInfo{
			ID:        msg.ID,
			ChatID:    msg.ChatID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
			IsGroup:   msg.IsGroup,
			HasMedia:  msg.HasMedia,
			MediaType: msg.MediaType,
		}
	}

	searchResult := SearchResult{
		Query:   searchFlags.query,
		Count:   len(results),
		Results: results,
	}

	if sendFlags.jsonOutput {
		output, _ := json.MarshalIndent(searchResult, "", "  ")
		fmt.Println(string(output))
	} else {
		printSearchResults(&searchResult)
	}

	return nil
}

func printSearchResults(result *SearchResult) {
	if result.Count == 0 {
		fmt.Println("No messages found matching your query.")
		return
	}

	fmt.Printf("Found %d message(s) matching: %s\n\n", result.Count, result.Query)

	for _, msg := range result.Results {
		prefix := "📱"
		if msg.IsGroup {
			prefix = "👥"
		}
		if msg.HasMedia {
			prefix = "📎"
		}

		timeStr := msg.Timestamp.Format("2006-01-02 15:04")
		content := msg.Content
		if len(content) > 80 {
			content = content[:77] + "..."
		}
		// Replace newlines for display
		content = strings.ReplaceAll(content, "\n", " ")

		fmt.Printf("%s [%s] %s\n", prefix, timeStr, msg.ChatID)
		fmt.Printf("   %s\n\n", content)
	}
}
