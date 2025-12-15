package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunEncrypt handles the encrypt command.
// It loads the user's identity, decrypts the vault key, and encrypts data from input to output.
func RunEncrypt(inputPath, outputPath string, cfg *config.Config) error {
	// Load and decrypt vault key
	vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load vault key: %w", err)
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

	// Encrypt data
	if err := vaultKey.Encrypt(input, output); err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	return nil
}
