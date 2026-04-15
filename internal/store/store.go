package store

import (
	"os"
	"path/filepath"
)

const DefaultStoreDir = ".wacli"

// GetStoreDir returns the store directory path
func GetStoreDir(customPath string) string {
	if customPath != "" {
		return customPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, DefaultStoreDir)
}

// EnsureStoreDir creates the store directory if it doesn't exist
func EnsureStoreDir(path string) error {
	return os.MkdirAll(path, 0700)
}
