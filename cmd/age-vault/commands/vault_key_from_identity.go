package commands

import (
	"fmt"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunVaultKeyFromIdentity handles the vault-key from-identity command.
// It creates a new vault key encrypted with the user's identity.
func RunVaultKeyFromIdentity(cfg *config.Config) error {
	fmt.Fprintf(os.Stderr, "[DEBUG] Starting vault-key from-identity\n")
	fmt.Fprintf(os.Stderr, "[DEBUG] VaultKeyFile: %s\n", cfg.VaultKeyFile)
	fmt.Fprintf(os.Stderr, "[DEBUG] IdentityFile: %s\n", cfg.IdentityFile)

	// Check if vault key file already exists
	fmt.Fprintf(os.Stderr, "[DEBUG] Checking if vault key file exists...\n")
	if _, err := os.Stat(cfg.VaultKeyFile); err == nil {
		return fmt.Errorf("vault key already exists at %s", cfg.VaultKeyFile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check vault key file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Vault key file does not exist, continuing...\n")

	// Load the user's identity from the configured identity file
	fmt.Fprintf(os.Stderr, "[DEBUG] Loading user identity...\n")
	userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] User identity loaded successfully. Type: %T\n", userIdentity)

	// Generate a new vault key
	fmt.Fprintf(os.Stderr, "[DEBUG] Generating new vault key...\n")
	vaultKeyIdentity, err := vault.GenerateVaultKey()
	if err != nil {
		return fmt.Errorf("failed to generate vault key: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Vault key generated successfully\n")

	// Extract the recipient (public key) from the loaded identity
	fmt.Fprintf(os.Stderr, "[DEBUG] Extracting recipient from user identity...\n")
	userRecipient, err := keymgmt.ExtractRecipient(userIdentity)
	if err != nil {
		return fmt.Errorf("failed to extract recipient from identity: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Recipient extracted successfully. Type: %T\n", userRecipient)

	// Encrypt the vault key for this recipient
	fmt.Fprintf(os.Stderr, "[DEBUG] Encrypting vault key for recipient...\n")
	encryptedVaultKey, err := vault.EncryptVaultKey(vaultKeyIdentity, userRecipient)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Vault key encrypted successfully. Size: %d bytes\n", len(encryptedVaultKey))

	// Ensure parent directory exists for the vault key file
	fmt.Fprintf(os.Stderr, "[DEBUG] Ensuring parent directory exists...\n")
	if err := config.EnsureParentDir(cfg.VaultKeyFile); err != nil {
		return err
	}

	// Write encrypted vault key to the configured location with secure permissions
	fmt.Fprintf(os.Stderr, "[DEBUG] Writing encrypted vault key to file...\n")
	if err := os.WriteFile(cfg.VaultKeyFile, encryptedVaultKey, 0600); err != nil {
		return fmt.Errorf("failed to save vault key: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Vault key saved successfully\n")
	fmt.Printf("Vault key successfully created at %s\n", cfg.VaultKeyFile)
	return nil
}
