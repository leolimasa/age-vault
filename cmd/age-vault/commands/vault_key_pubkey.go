package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunVaultKeyPubkey handles the vault-key pubkey command.
// It extracts and outputs the public key from the vault key.
func RunVaultKeyPubkey(outputPath string, cfg *config.Config) error {
	// Load and decrypt vault key
	vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load vault key: %w", err)
	}

	// Get the identity from the vault key
	vaultKeyIdentity, err := vaultKey.GetIdentity()
	if err != nil {
		return fmt.Errorf("failed to get vault key identity: %w", err)
	}

	// Extract the recipient (public key) from the vault key identity
	recipient, err := keymgmt.ExtractRecipient(vaultKeyIdentity)
	if err != nil {
		return fmt.Errorf("failed to extract recipient from vault key: %w", err)
	}

	// Get the string representation of the public key
	pubKeyStr, err := keymgmt.RecipientToString(recipient)
	if err != nil {
		return fmt.Errorf("failed to convert recipient to string: %w", err)
	}
	pubKeyStr = pubKeyStr + "\n"

	// Open output (file or stdout)
	var output io.Writer
	if outputPath == "" {
		output = os.Stdout
	} else {
		if err := config.EnsureParentDir(outputPath); err != nil {
			return err
		}
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		output = f
	}

	// Write the public key
	if _, err := io.WriteString(output, pubKeyStr); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}
