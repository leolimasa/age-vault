package commands

import (
	"bytes"
	"fmt"
	"os"

	"filippo.io/age"
	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunVaultKeyEncrypt handles the vault-key encrypt command.
// It creates a new vault key if none exists, or loads the existing one,
// then encrypts it for a new recipient (user's public key).
// Supports two ways to specify the recipient: --pubkey or --pubkey-file.
// Outputs to stdout by default, unless --save or -o is specified.
func RunVaultKeyEncrypt(pubkey string, pubkeyFile string, outputPath string, save bool, cfg *config.Config) error {
	// Validate that exactly one recipient source is provided
	providedCount := 0
	if pubkey != "" {
		providedCount++
	}
	if pubkeyFile != "" {
		providedCount++
	}
	if providedCount != 1 {
		return fmt.Errorf("exactly one of --pubkey or --pubkey-file must be provided")
	}
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
		userRecipient, err := keymgmt.ExtractRecipient(userIdentity)
		if err != nil {
			return fmt.Errorf("failed to extract recipient from identity: %w", err)
		}

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
		// Vault key exists, load and decrypt it using helper function
		vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load vault key: %w", err)
		}

		// Get the identity from the vault key
		vaultKeyIdentity, err = vaultKey.GetIdentity()
		if err != nil {
			return fmt.Errorf("failed to get vault key identity: %w", err)
		}
	}

	// Get the recipient based on which flag was provided
	var recipient age.Recipient

	if pubkey != "" {
		// Parse recipient from string
		recipients, parseErr := age.ParseRecipients(bytes.NewReader([]byte(pubkey)))
		if parseErr != nil {
			return fmt.Errorf("failed to parse public key: %w", parseErr)
		}
		if len(recipients) == 0 {
			return fmt.Errorf("no recipient found in provided public key")
		}
		recipient = recipients[0]
	} else if pubkeyFile != "" {
		// Read and parse recipient from file
		content, readErr := os.ReadFile(pubkeyFile)
		if readErr != nil {
			return fmt.Errorf("failed to read public key file: %w", readErr)
		}
		recipients, parseErr := age.ParseRecipients(bytes.NewReader(content))
		if parseErr != nil {
			return fmt.Errorf("failed to parse public key file: %w", parseErr)
		}
		if len(recipients) == 0 {
			return fmt.Errorf("no recipient found in public key file")
		}
		recipient = recipients[0]
	}

	// Encrypt vault key for the recipient
	encryptedKey, err := vault.EncryptVaultKey(vaultKeyIdentity, recipient)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key for recipient: %w", err)
	}

	// Determine output destination
	if outputPath != "" {
		// Write to specified output file
		if err := config.EnsureParentDir(outputPath); err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, encryptedKey, 0600); err != nil {
			return fmt.Errorf("failed to write encrypted vault key: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Vault key encrypted and saved to %s\n", outputPath)
	} else if save {
		// Save to config vault key file
		if err := config.EnsureParentDir(cfg.VaultKeyFile); err != nil {
			return err
		}
		if err := os.WriteFile(cfg.VaultKeyFile, encryptedKey, 0600); err != nil {
			return fmt.Errorf("failed to write encrypted vault key: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Vault key encrypted and saved to %s\n", cfg.VaultKeyFile)
	} else {
		// Output to stdout
		if _, err := os.Stdout.Write(encryptedKey); err != nil {
			return fmt.Errorf("failed to write encrypted vault key to stdout: %w", err)
		}
	}

	return nil
}
