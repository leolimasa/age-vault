package commands

import (
	"filippo.io/age"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/vault"
)

func TestRunVaultKeyPubkey(t *testing.T) {
	tempDir := t.TempDir()

	// Generate a test identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	// Write identity to file
	identityPath := filepath.Join(tempDir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0600); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Generate a vault key
	vaultKeyIdentity, err := vault.GenerateVaultKey()
	if err != nil {
		t.Fatalf("failed to generate vault key: %v", err)
	}

	// Get the expected public key
	vaultX25519, ok := vaultKeyIdentity.(*age.X25519Identity)
	if !ok {
		t.Fatal("vault key is not an X25519Identity")
	}
	expectedPubkey := vaultX25519.Recipient().String()

	// Encrypt vault key for our identity
	recipient := identity.Recipient()
	encryptedKey, err := vault.EncryptVaultKey(vaultKeyIdentity, recipient)
	if err != nil {
		t.Fatalf("failed to encrypt vault key: %v", err)
	}

	// Write encrypted vault key to file
	vaultKeyPath := filepath.Join(tempDir, "vault_key.age")
	if err := os.WriteFile(vaultKeyPath, encryptedKey, 0600); err != nil {
		t.Fatalf("failed to write vault key file: %v", err)
	}

	// Create config
	cfg := &config.Config{
		IdentityFile: identityPath,
		VaultKeyFile: vaultKeyPath,
	}

	// Test: Extract public key to stdout (simulated with temp file)
	outputPath := filepath.Join(tempDir, "pubkey.txt")
	err = RunVaultKeyPubkey(outputPath, cfg)
	if err != nil {
		t.Fatalf("RunVaultKeyPubkey failed: %v", err)
	}

	// Verify output
	pubkeyData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	pubkeyStr := strings.TrimSpace(string(pubkeyData))
	if pubkeyStr != expectedPubkey {
		t.Errorf("public key mismatch:\ngot:  %s\nwant: %s", pubkeyStr, expectedPubkey)
	}
}

func TestRunVaultKeyPubkey_MissingVaultKey(t *testing.T) {
	tempDir := t.TempDir()

	// Generate a test identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate test identity: %v", err)
	}

	identityPath := filepath.Join(tempDir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0600); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	cfg := &config.Config{
		IdentityFile: identityPath,
		VaultKeyFile: filepath.Join(tempDir, "nonexistent.age"),
	}

	outputPath := filepath.Join(tempDir, "pubkey.txt")
	err = RunVaultKeyPubkey(outputPath, cfg)
	if err == nil {
		t.Fatal("expected error with missing vault key file, got nil")
	}
}

func TestRunVaultKeyPubkey_WrongIdentity(t *testing.T) {
	tempDir := t.TempDir()

	// Generate two different identities
	identity1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity1: %v", err)
	}

	identity2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity2: %v", err)
	}

	// Write identity2 to file (we'll encrypt for identity1)
	identityPath := filepath.Join(tempDir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity2.String()+"\n"), 0600); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Generate vault key and encrypt for identity1 (not identity2)
	vaultKeyIdentity, err := vault.GenerateVaultKey()
	if err != nil {
		t.Fatalf("failed to generate vault key: %v", err)
	}

	recipient1 := identity1.Recipient()
	encryptedKey, err := vault.EncryptVaultKey(vaultKeyIdentity, recipient1)
	if err != nil {
		t.Fatalf("failed to encrypt vault key: %v", err)
	}

	vaultKeyPath := filepath.Join(tempDir, "vault_key.age")
	if err := os.WriteFile(vaultKeyPath, encryptedKey, 0600); err != nil {
		t.Fatalf("failed to write vault key file: %v", err)
	}

	// Create config with identity2 (wrong identity)
	cfg := &config.Config{
		IdentityFile: identityPath,
		VaultKeyFile: vaultKeyPath,
	}

	outputPath := filepath.Join(tempDir, "pubkey.txt")
	err = RunVaultKeyPubkey(outputPath, cfg)
	if err == nil {
		t.Fatal("expected error when using wrong identity to decrypt vault key, got nil")
	}
}
