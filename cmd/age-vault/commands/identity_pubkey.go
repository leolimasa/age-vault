package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunIdentityPubkey handles the identity pubkey command.
// It extracts and outputs the public key from the configured identity file.
func RunIdentityPubkey(outputPath string, cfg *config.Config) error {
	// Extract the public key string from the identity file
	pubKeyStr, err := keymgmt.ExtractRecipientString(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to extract public key from identity: %w", err)
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
