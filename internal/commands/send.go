package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send messages or files",
	Long:  `Send text messages or files to WhatsApp contacts or groups`,
}

var sendTextCmd = &cobra.Command{
	Use:   "text",
	Short: "Send a text message",
	Long:  `Send a text message to a contact or group`,
	RunE:  runSendText,
}

var sendFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Send a file",
	Long:  `Send a file with optional caption to a contact or group`,
	RunE:  runSendFile,
}

// sendFlags holds the flags for send commands
var sendFlags = struct {
	to        string
	message   string
	file      string
	caption   string
	jsonOutput bool
}{}

func init() {
	// send text command flags
	sendTextCmd.Flags().StringVarP(&sendFlags.to, "to", "t", "", "Recipient JID or phone number (required)")
	sendTextCmd.Flags().StringVarP(&sendFlags.message, "message", "m", "", "Message content (required)")
	sendTextCmd.Flags().BoolVar(&sendFlags.jsonOutput, "json", false, "Output in JSON format")

	sendTextCmd.MarkFlagRequired("to")
	sendTextCmd.MarkFlagRequired("message")

	// send file command flags
	sendFileCmd.Flags().StringVarP(&sendFlags.to, "to", "t", "", "Recipient JID or phone number (required)")
	sendFileCmd.Flags().StringVarP(&sendFlags.file, "file", "f", "", "Path to file to send (required)")
	sendFileCmd.Flags().StringVarP(&sendFlags.caption, "caption", "c", "", "Caption for the file")
	sendFileCmd.Flags().BoolVar(&sendFlags.jsonOutput, "json", false, "Output in JSON format")

	sendFileCmd.MarkFlagRequired("to")
	sendFileCmd.MarkFlagRequired("file")

	// Add subcommands to send
	sendCmd.AddCommand(sendTextCmd)
	sendCmd.AddCommand(sendFileCmd)
}

// SendResult represents the result of a send operation
type SendResult struct {
	Success bool   `json:"success"`
	To      string `json:"to"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func runSendText(cmd *cobra.Command, args []string) error {
	waClient := GetClient()
	if waClient == nil {
		return fmt.Errorf("client not initialized. Run 'wacli auth' first")
	}

	if !waClient.IsConnected() {
		return fmt.Errorf("client not connected. Run 'wacli auth' first")
	}

	ctx := context.Background()
	err := waClient.SendTextMessage(ctx, sendFlags.to, sendFlags.message)

	result := SendResult{
		To:      sendFlags.to,
		Message: sendFlags.message,
		Success: err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}

	if sendFlags.jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		if result.Success {
			fmt.Printf("✓ Message sent to %s\n", sendFlags.to)
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to send message: %s\n", result.Error)
		}
	}

	return err
}

func runSendFile(cmd *cobra.Command, args []string) error {
	waClient := GetClient()
	if waClient == nil {
		return fmt.Errorf("client not initialized. Run 'wacli auth' first")
	}

	if !waClient.IsConnected() {
		return fmt.Errorf("client not connected. Run 'wacli auth' first")
	}

	// Check if file exists
	if _, err := os.Stat(sendFlags.file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", sendFlags.file)
	}

	ctx := context.Background()
	err := waClient.SendFile(ctx, sendFlags.to, sendFlags.file, sendFlags.caption)

	result := SendResult{
		To:      sendFlags.to,
		Message: sendFlags.caption,
		Success: err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}

	if sendFlags.jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		if result.Success {
			fmt.Printf("✓ File sent to %s\n", sendFlags.to)
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to send file: %s\n", result.Error)
		}
	}

	return err
}
