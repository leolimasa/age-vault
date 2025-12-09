package commands

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/vault"
)

func TestRunVaultKeyFromIdentity(t *testing.T) {
	// Create temporary directory for test
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

	// Create config
	vaultKeyPath := filepath.Join(tempDir, "vault_key.age")
	cfg := &config.Config{
		IdentityFile: identityPath,
		VaultKeyFile: vaultKeyPath,
	}

	// Test: Create vault key
	err = RunVaultKeyFromIdentity(cfg)
	if err != nil {
		t.Fatalf("RunVaultKeyFromIdentity failed: %v", err)
	}

	// Verify vault key file was created
	if _, err := os.Stat(vaultKeyPath); os.IsNotExist(err) {
		t.Fatal("vault key file was not created")
	}

	// Verify file permissions
	info, err := os.Stat(vaultKeyPath)
	if err != nil {
		t.Fatalf("failed to stat vault key file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("vault key file has wrong permissions: got %o, want 0600", info.Mode().Perm())
	}

	// Verify we can decrypt the vault key
	encryptedKey, err := os.ReadFile(vaultKeyPath)
	if err != nil {
		t.Fatalf("failed to read vault key file: %v", err)
	}

	_, err = vault.DecryptVaultKey(encryptedKey, identity)
	if err != nil {
		t.Fatalf("failed to decrypt vault key: %v", err)
	}

	// Test: Should fail on second run (vault key already exists)
	err = RunVaultKeyFromIdentity(cfg)
	if err == nil {
		t.Fatal("expected error when vault key already exists, got nil")
	}
}

func TestRunVaultKeyFromIdentity_MissingIdentity(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		IdentityFile: filepath.Join(tempDir, "nonexistent.txt"),
		VaultKeyFile: filepath.Join(tempDir, "vault_key.age"),
	}

	err := RunVaultKeyFromIdentity(cfg)
	if err == nil {
		t.Fatal("expected error with missing identity file, got nil")
	}
}

func TestRunVaultKeyFromIdentity_CreatesParentDir(t *testing.T) {
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

	// Use a vault key path with non-existent parent directory
	vaultKeyPath := filepath.Join(tempDir, "subdir", "nested", "vault_key.age")
	cfg := &config.Config{
		IdentityFile: identityPath,
		VaultKeyFile: vaultKeyPath,
	}

	err = RunVaultKeyFromIdentity(cfg)
	if err != nil {
		t.Fatalf("RunVaultKeyFromIdentity failed: %v", err)
	}

	// Verify vault key file was created
	if _, err := os.Stat(vaultKeyPath); os.IsNotExist(err) {
		t.Fatal("vault key file was not created in nested directory")
	}
}
