package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("AGE_VAULT_KEY_FILE")
	os.Unsetenv("AGE_VAULT_IDENTITY_FILE")
	os.Unsetenv("AGE_VAULT_SSH_KEYS_DIR")

	// Change to a temp directory to avoid finding any existing age_vault.yml
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedConfigDir := filepath.Join(homeDir, ".config", ".age-vault")

	if cfg.VaultKeyFile != filepath.Join(expectedConfigDir, "vault_key.age") {
		t.Errorf("Expected VaultKeyFile to be %s, got %s", filepath.Join(expectedConfigDir, "vault_key.age"), cfg.VaultKeyFile)
	}

	if cfg.IdentityFile != filepath.Join(expectedConfigDir, "identity.txt") {
		t.Errorf("Expected IdentityFile to be %s, got %s", filepath.Join(expectedConfigDir, "identity.txt"), cfg.IdentityFile)
	}

	if cfg.SSHKeysDir != "" {
		t.Errorf("Expected SSHKeysDir to be empty, got %s", cfg.SSHKeysDir)
	}
}

func TestNewConfig_EnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("AGE_VAULT_KEY_FILE", "/custom/vault_key.age")
	os.Setenv("AGE_VAULT_IDENTITY_FILE", "/custom/identity.txt")
	os.Setenv("AGE_VAULT_SSH_KEYS_DIR", "/custom/ssh_keys")
	defer func() {
		os.Unsetenv("AGE_VAULT_KEY_FILE")
		os.Unsetenv("AGE_VAULT_IDENTITY_FILE")
		os.Unsetenv("AGE_VAULT_SSH_KEYS_DIR")
	}()

	// Change to a temp directory to avoid finding any existing age_vault.yml
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	if cfg.VaultKeyFile != "/custom/vault_key.age" {
		t.Errorf("Expected VaultKeyFile to be /custom/vault_key.age, got %s", cfg.VaultKeyFile)
	}

	if cfg.IdentityFile != "/custom/identity.txt" {
		t.Errorf("Expected IdentityFile to be /custom/identity.txt, got %s", cfg.IdentityFile)
	}

	if cfg.SSHKeysDir != "/custom/ssh_keys" {
		t.Errorf("Expected SSHKeysDir to be /custom/ssh_keys, got %s", cfg.SSHKeysDir)
	}
}

func TestExpandHomePath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(homeDir, "test")},
		{"~", homeDir},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		result := expandHomePath(tt.input)
		if result != tt.expected {
			t.Errorf("expandHomePath(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestEnsureParentDir(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "subdir", "nested", "file.txt")

	err := EnsureParentDir(testFile)
	if err != nil {
		t.Fatalf("EnsureParentDir() failed: %v", err)
	}

	// Check that the parent directory exists
	parentDir := filepath.Dir(testFile)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent directory %s was not created", parentDir)
	}

	// Check permissions
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("Error stating directory: %v", err)
	}

	// On Unix systems, check that permissions are 0700
	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected directory permissions to be 0700, got %o", info.Mode().Perm())
	}
}
