package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/keymgmt"
	"github.com/leolimasa/age-vault/sshagent"
	"github.com/leolimasa/age-vault/vault"
)

// RunSSHStartAgent handles the ssh start-agent command.
// It starts an SSH agent that manages vault-encrypted SSH keys.
func RunSSHStartAgent(keysDir string, cfg *config.Config) error {
	// Use provided keysDir or fall back to config
	if keysDir == "" {
		keysDir = cfg.SSHKeysDir
	}

	if keysDir == "" {
		return fmt.Errorf("SSH keys directory not specified (use --keys-dir or set AGE_VAULT_SSH_KEYS_DIR)")
	}

	// Check if keys directory exists
	if _, err := os.Stat(keysDir); os.IsNotExist(err) {
		return fmt.Errorf("SSH keys directory does not exist: %s", keysDir)
	}

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

	// Create VaultSSHAgent
	agent, err := sshagent.NewVaultSSHAgent(keysDir, vaultKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH agent: %w", err)
	}

	// Determine socket path
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("age-vault-ssh-agent-%d.sock", os.Getpid()))

	// Start the agent
	if err := agent.Start(socketPath); err != nil {
		return fmt.Errorf("failed to start SSH agent: %w", err)
	}

	return nil
}
