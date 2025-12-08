package commands

import (
	"filippo.io/age"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/vault"
)

// RunSops handles the sops passthrough command.
// It decrypts the vault key and makes it available to sops via environment variable,
// then executes sops with the provided arguments.
func RunSops(sopsArgs []string, cfg *config.Config) error {
	// Load user's identity
	userIdentity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return fmt.Errorf("failed to load identity: %w", err)
	}

	// Read encrypted vault key
	encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read vault key file: %w", err)
	}

	// Decrypt vault key
	vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
	if err != nil {
		return fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	// Get the vault key identity
	vaultKeyIdentity, err := vaultKey.GetIdentity()
	if err != nil {
		return fmt.Errorf("failed to get vault key identity: %w", err)
	}

	// Convert to X25519Identity to get the string representation
	x25519Identity, ok := vaultKeyIdentity.(*age.X25519Identity)
	if !ok {
		return fmt.Errorf("vault key must be an X25519Identity")
	}
	vaultKeyStr := x25519Identity.String()

	// Create a temporary file to hold the vault key
	// (sops needs a file path, we can't use a pipe directly)
	tempDir, err := os.MkdirTemp("", "age-vault-sops-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after sops finishes

	keyFile := filepath.Join(tempDir, "key.txt")
	if err := os.WriteFile(keyFile, []byte(vaultKeyStr), 0600); err != nil {
		return fmt.Errorf("failed to write vault key to temp file: %w", err)
	}

	// Set SOPS_AGE_KEY_FILE environment variable
	cmd := exec.Command("sops", sopsArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run sops
	if err := cmd.Run(); err != nil {
		// Check if the error is because sops is not found
		if exitErr, ok := err.(*exec.ExitError); ok {
			// sops returned a non-zero exit code
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to run sops: %w", err)
	}

	return nil
}
