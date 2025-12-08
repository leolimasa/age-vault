package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leolimasa/age-vault/config"
)

// RunSSHListKeys handles the ssh list-keys command.
// It lists all encrypted SSH keys in the configured directory.
func RunSSHListKeys(cfg *config.Config) error {
	if cfg.SSHKeysDir == "" {
		return fmt.Errorf("SSH keys directory not specified (set AGE_VAULT_SSH_KEYS_DIR)")
	}

	// Check if directory exists
	if _, err := os.Stat(cfg.SSHKeysDir); os.IsNotExist(err) {
		return fmt.Errorf("SSH keys directory does not exist: %s", cfg.SSHKeysDir)
	}

	// List all .age files
	matches, err := filepath.Glob(filepath.Join(cfg.SSHKeysDir, "*.age"))
	if err != nil {
		return fmt.Errorf("error scanning keys directory: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No encrypted SSH keys found")
		return nil
	}

	fmt.Printf("Found %d encrypted SSH key(s) in %s:\n", len(matches), cfg.SSHKeysDir)
	for _, keyPath := range matches {
		fmt.Printf("  - %s\n", filepath.Base(keyPath))
	}

	return nil
}
