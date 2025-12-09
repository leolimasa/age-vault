package commands

import (
	"filippo.io/age"
	"fmt"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunVaultKeyFromIdentity handles the vault-key from-identity command.
// It creates a new vault key encrypted with the user's identity.
func RunVaultKeyFromIdentity(cfg *config.Config) error {
	// Check if vault key file already exists
	if _, err := os.Stat(cfg.VaultKeyFile); err == nil {
		return fmt.Errorf("vault key already exists at %s", cfg.VaultKeyFile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check vault key file: %w", err)
	}

	// Load the user's identity from the configured identity file
	userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}

	// Generate a new vault key
	vaultKeyIdentity, err := vault.GenerateVaultKey()
	if err != nil {
		return fmt.Errorf("failed to generate vault key: %w", err)
	}

	// Extract the recipient (public key) from the loaded identity
	userX25519, ok := userIdentity.(*age.X25519Identity)
	if !ok {
		return fmt.Errorf("identity must be an X25519Identity")
	}
	userRecipient := userX25519.Recipient()

	// Encrypt the vault key for this recipient
	encryptedVaultKey, err := vault.EncryptVaultKey(vaultKeyIdentity, userRecipient)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key: %w", err)
	}

	// Ensure parent directory exists for the vault key file
	if err := config.EnsureParentDir(cfg.VaultKeyFile); err != nil {
		return err
	}

	// Write encrypted vault key to the configured location with secure permissions
	if err := os.WriteFile(cfg.VaultKeyFile, encryptedVaultKey, 0600); err != nil {
		return fmt.Errorf("failed to save vault key: %w", err)
	}

	fmt.Printf("Vault key successfully created at %s\n", cfg.VaultKeyFile)
	return nil
}
