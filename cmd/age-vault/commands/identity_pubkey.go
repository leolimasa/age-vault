package commands

import (
	"filippo.io/age"
	"fmt"
	"io"
	"os"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
)

// RunIdentityPubkey handles the identity pubkey command.
// It extracts and outputs the public key from the configured identity file.
func RunIdentityPubkey(outputPath string, cfg *config.Config) error {
	// Load the identity
	identity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}

	// Get the public key (recipient) from the identity
	x25519Identity, ok := identity.(*age.X25519Identity)
	if !ok {
		return fmt.Errorf("identity must be an X25519Identity")
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
