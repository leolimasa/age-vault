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

// RunVaultKeyEncrypt handles the vault-key encrypt command.
// It creates a new vault key if none exists, or loads the existing one,
// then encrypts it for a new recipient (user's public key).
func RunVaultKeyEncrypt(pubKeyPath, outputPath string, cfg *config.Config) error {
	var vaultKeyIdentity age.Identity

	// Check if vault key exists
	if _, err := os.Stat(cfg.VaultKeyFile); os.IsNotExist(err) {
		// Vault key doesn't exist, generate a new one
		identity, err := vault.GenerateVaultKey()
		if err != nil {
			return fmt.Errorf("failed to generate vault key: %w", err)
		}

		// Save the vault key encrypted for ourselves first
		// Load our identity
		userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to load identity: %w", err)
		}

		// Get our recipient (public key) from our identity
		userX25519, ok := userIdentity.(*age.X25519Identity)
		if !ok {
			return fmt.Errorf("identity must be an X25519Identity")
		}
		userRecipient := userX25519.Recipient()

		// Encrypt vault key for ourselves
		encryptedForUs, err := vault.EncryptVaultKey(identity, userRecipient)
		if err != nil {
			return fmt.Errorf("failed to encrypt vault key for self: %w", err)
		}

		// Ensure parent directory exists
		if err := config.EnsureParentDir(cfg.VaultKeyFile); err != nil {
			return err
		}

		// Save the encrypted vault key
		if err := os.WriteFile(cfg.VaultKeyFile, encryptedForUs, 0600); err != nil {
			return fmt.Errorf("failed to save vault key: %w", err)
		}

		vaultKeyIdentity = identity
	} else {
		// Vault key exists, load and decrypt it
		userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
		if err != nil {
			return fmt.Errorf("failed to load identity: %w", err)
		}

		encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read vault key file: %w", err)
		}

		vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
		if err != nil {
			return fmt.Errorf("failed to decrypt vault key: %w", err)
		}

		// Get the identity from the vault key using a helper method
		vaultKeyIdentity, err = vaultKey.GetIdentity()
		if err != nil {
			return fmt.Errorf("failed to get vault key identity: %w", err)
		}
	}

	// Load the recipient's public key
	recipient, err := keymgmt.LoadRecipient(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load recipient public key: %w", err)
	}

	// Encrypt vault key for the recipient
	encryptedKey, err := vault.EncryptVaultKey(vaultKeyIdentity, recipient)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key for recipient: %w", err)
	}

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

	// Write encrypted vault key
	if _, err := output.Write(encryptedKey); err != nil {
		return fmt.Errorf("failed to write encrypted vault key: %w", err)
	}

	return nil
}
