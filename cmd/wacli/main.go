package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wacli",
	Short: "WhatsApp CLI - A terminal client for WhatsApp",
	Long: `wacli is a command-line interface for WhatsApp, built on top of whatsmeow.
It supports message sync, search, sending, and group management.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WhatsApp CLI - Use --help for available commands")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
