// Package config provides configuration management for age-vault.
// It supports environment variables, YAML config files, and default values.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for age-vault.
type Config struct {
	VaultKeyFile  string // Path to encrypted vault key
	IdentityFile  string // Path to user's private key (identity)
	SSHKeysDir    string // Directory containing encrypted SSH keys
	configFileDir string // Directory containing the loaded config file (private)
}

// yamlConfig represents the structure of age_vault.yml file.
type yamlConfig struct {
	VaultKeyFile string `yaml:"vault_key_file"`
	IdentityFile string `yaml:"identity_file"`
	SSHKeysDir   string `yaml:"ssh_keys_dir"`
}

// NewConfig creates a new Config by reading environment variables,
// searching for age_vault.yml config file, and applying defaults.
// Environment variables override YAML config, which overrides defaults.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// First, load from YAML config file if present
	yamlCfg, configFileDir, err := findAndLoadYAMLConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading YAML config: %w", err)
	}

	// Store the config file directory
	cfg.configFileDir = configFileDir

	// Get home directory for defaults
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	defaultConfigDir := filepath.Join(homeDir, ".config", ".age-vault")

	// Resolve YAML paths relative to config file directory
	resolvedVaultKeyFile := resolveConfigPath(yamlCfg.VaultKeyFile, configFileDir)
	resolvedIdentityFile := resolveConfigPath(yamlCfg.IdentityFile, configFileDir)
	resolvedSSHKeysDir := resolveConfigPath(yamlCfg.SSHKeysDir, configFileDir)

	// Set VaultKeyFile
	cfg.VaultKeyFile = getConfigValue(
		os.Getenv("AGE_VAULT_KEY_FILE"),
		resolvedVaultKeyFile,
		filepath.Join(defaultConfigDir, "vault_key.age"),
	)

	// Set IdentityFile
	cfg.IdentityFile = getConfigValue(
		os.Getenv("AGE_VAULT_IDENTITY_FILE"),
		resolvedIdentityFile,
		filepath.Join(defaultConfigDir, "identity.txt"),
	)

	// Set SSHKeysDir (no default)
	cfg.SSHKeysDir = getConfigValue(
		os.Getenv("AGE_VAULT_SSH_KEYS_DIR"),
		resolvedSSHKeysDir,
		"",
	)

	// Expand home directory in all paths (only for env vars or defaults)
	cfg.VaultKeyFile = expandHomePath(cfg.VaultKeyFile)
	cfg.IdentityFile = expandHomePath(cfg.IdentityFile)
	if cfg.SSHKeysDir != "" {
		cfg.SSHKeysDir = expandHomePath(cfg.SSHKeysDir)
	}

	return cfg, nil
}

// findAndLoadYAMLConfig searches for age_vault.yml by traversing up the directory tree.
// Returns the config, the directory containing the config file, and any error.
func findAndLoadYAMLConfig() (yamlConfig, string, error) {
	cfg := yamlConfig{}

	// Start from current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return cfg, "", nil // Not fatal, just use defaults
	}

	// Traverse up the directory tree
	for {
		configPath := filepath.Join(currentDir, "age_vault.yml")
		if _, err := os.Stat(configPath); err == nil {
			// File exists, try to load it
			data, err := os.ReadFile(configPath)
			if err != nil {
				return cfg, "", fmt.Errorf("error reading config file %s: %w", configPath, err)
			}

			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return cfg, "", fmt.Errorf("error parsing config file %s: %w", configPath, err)
			}

			return cfg, currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return cfg, "", nil // No config file found, return empty config
}

// getConfigValue returns the first non-empty value from the arguments.
func getConfigValue(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// expandHomePath expands ~ in path to the user's home directory.
func expandHomePath(path string) string {
	if path == "" {
		return path
	}

	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path // Return unchanged if we can't get home dir
		}

		if len(path) == 1 {
			return homeDir
		}

		if path[1] == filepath.Separator {
			return filepath.Join(homeDir, path[2:])
		}
	}

	return path
}

// resolveConfigPath resolves a path from a config file.
// - If path is empty, returns empty string
// - If path is absolute, returns as-is
// - If path starts with ~, expands home directory
// - Otherwise, treats as relative path and joins with configFileDir
func resolveConfigPath(path string, configFileDir string) string {
	if path == "" {
		return ""
	}

	// Check if it's an absolute path
	if filepath.IsAbs(path) {
		return path
	}

	// Check if it starts with ~
	if len(path) > 0 && path[0] == '~' {
		return expandHomePath(path)
	}

	// If we have a config file directory, resolve relative to it
	if configFileDir != "" {
		return filepath.Join(configFileDir, path)
	}

	// Otherwise, return as-is
	return path
}

// EnsureParentDir creates the parent directory for the given file path if it doesn't exist.
// Sets permissions to 0700 for security.
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}
	return nil
}
