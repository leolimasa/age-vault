package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunDecrypt handles the decrypt command.
// It loads the user's identity, decrypts the vault key, and decrypts data from input to output.
func RunDecrypt(inputPath, outputPath string, cfg *config.Config) error {
	// Load user's identity
	userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}

	// Read encrypted vault key
	// REVIEW this should be a common function to read and decrypt vault keys
	encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read vault key file: %w", err)
	}

	// Decrypt vault key
	vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
	if err != nil {
		return fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	// Open input (file or stdin)
	var input io.Reader
	if inputPath == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer f.Close()
		input = f
	}

	// Open output (file or stdout)
	var output io.Writer
	if outputPath == "" {
		output = os.Stdout
	} else {
		// Ensure parent directory exists
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

	// Decrypt data
	if err := vaultKey.Decrypt(input, output); err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	return nil
}
