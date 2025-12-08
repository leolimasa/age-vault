package commands

import (
	"fmt"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunIdentitySet handles the identity set command.
// It copies the provided identity file to the configured identity location.
func RunIdentitySet(sourcePath string, cfg *config.Config) error {
	if err := keymgmt.CopyFile(sourcePath, cfg.IdentityFile); err != nil {
		return fmt.Errorf("failed to set identity: %w", err)
	}

	fmt.Printf("Identity set to: %s\n", cfg.IdentityFile)
	return nil
}
