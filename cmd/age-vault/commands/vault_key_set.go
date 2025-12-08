package commands

import (
	"fmt"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunVaultKeySet handles the vault-key set command.
// It copies the provided encrypted vault key file to the configured vault key location.
func RunVaultKeySet(sourcePath string, cfg *config.Config) error {
	if err := keymgmt.CopyFile(sourcePath, cfg.VaultKeyFile); err != nil {
		return fmt.Errorf("failed to set vault key: %w", err)
	}

	fmt.Printf("Vault key set to: %s\n", cfg.VaultKeyFile)
	return nil
}
