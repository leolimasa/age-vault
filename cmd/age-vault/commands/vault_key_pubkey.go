package commands

import (
	"filippo.io/age"
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunVaultKeyPubkey handles the vault-key pubkey command.
// It extracts and outputs the public key from the vault key.
func RunVaultKeyPubkey(outputPath string, cfg *config.Config) error {
	// Load the user's identity
	userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}

	// Read the encrypted vault key
	encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read vault key file: %w", err)
	}

	// Decrypt the vault key
	vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
	if err != nil {
		return fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	// Get the identity from the vault key
	vaultKeyIdentity, err := vaultKey.GetIdentity()
	if err != nil {
		return fmt.Errorf("failed to get vault key identity: %w", err)
	}

	// Get the public key (recipient) from the vault key identity
	x25519Identity, ok := vaultKeyIdentity.(*age.X25519Identity)
	if !ok {
		return fmt.Errorf("vault key identity must be an X25519Identity")
	}
	recipient := x25519Identity.Recipient()

	// Get the string representation of the public key
	pubKeyStr := recipient.String() + "\n"

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
