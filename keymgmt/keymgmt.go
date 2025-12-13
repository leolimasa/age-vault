// Package keymgmt provides key management utilities for age-vault.
// It handles loading identities, recipients, and secure file operations.
package keymgmt

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"filippo.io/age/plugin"

	// REVIEW: why are you using fully qualified domain names for local files??
	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/vault"
)

// LoadIdentity reads and parses an age identity file from the given path.
// Supports both native X25519 identities and plugin-based identities.
// Returns the first identity found in the file.
func LoadIdentity(path string) (age.Identity, error) {
	// Read the identity file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading identity file %s: %w", path, err)
	}

	// Try parsing as native age identity first
	identities, err := age.ParseIdentities(bytes.NewReader(content))
	if err == nil && len(identities) > 0 {
		return identities[0], nil
	}

	// If that failed, try parsing as plugin identity
	// Plugin identities have the format: AGE-PLUGIN-{NAME}-{DATA}
	// Need to extract just the identity line (skip comments starting with #)
	var identityStr string
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		// Skip empty lines and comments
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}
		// Found the identity line
		if bytes.HasPrefix(trimmed, []byte("AGE-PLUGIN-")) {
			identityStr = string(trimmed)
			break
		}
	}

	if identityStr == "" {
		return nil, fmt.Errorf("no identity found in file %s", path)
	}

	// We need to create a ClientUI for the plugin
	// REVIEW: you are recreating this several times on the same file. Have it share the same client UI
	ui := &plugin.ClientUI{
		DisplayMessage: func(name, message string) error {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] %s\n", name, message)
			return nil
		},
		RequestValue: func(name, prompt string, secret bool) (string, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] RequestValue called: %s (secret=%v)\n", name, prompt, secret)
			// Return empty string for non-secret prompts, error for secret ones
			if secret {
				return "", fmt.Errorf("cannot provide secret values in non-interactive mode")
			}
			return "", nil
		},
		Confirm: func(name, prompt, yes, no string) (bool, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] Confirm called: %s (yes=%s, no=%s)\n", name, prompt, yes, no)
			// Auto-confirm to proceed without user interaction
			// REVIEW: Don't just auto confirm. Prompte the user.
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] Auto-confirming (returning yes)\n", name)
			return true, nil
		},
		WaitTimer: func(name string) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] WaitTimer called - plugin is waiting...\n", name)
		},
	}

	pluginIdentity, pluginErr := plugin.NewIdentity(identityStr, ui)
	if pluginErr == nil {
		return pluginIdentity, nil
	}

	// If both failed, return an appropriate error
	return nil, fmt.Errorf("error parsing identity file %s as native identity: %w, as plugin identity: %v", path, err, pluginErr)
}

// LoadRecipient reads and parses an age public key file from the given path.
// Supports both native X25519 recipients and plugin-based recipients.
// Returns the first recipient found in the file.
func LoadRecipient(path string) (age.Recipient, error) {
	// Read the recipient file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading recipient file %s: %w", path, err)
	}

	// Try parsing as native age recipients first
	recipients, err := age.ParseRecipients(bytes.NewReader(content))
	if err == nil && len(recipients) > 0 {
		return recipients[0], nil
	}

	// If that failed, try parsing as plugin recipient
	// Plugin recipients have the format: age1{name}{data}
	// Need to extract just the recipient line (skip comments starting with #)
	var recipientStr string
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		// Skip empty lines and comments
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}
		// Found a recipient line (starts with age1)
		if bytes.HasPrefix(trimmed, []byte("age1")) {
			recipientStr = string(trimmed)
			break
		}
	}

	if recipientStr == "" {
		return nil, fmt.Errorf("no recipient found in file %s", path)
	}

	ui := &plugin.ClientUI{
		DisplayMessage: func(name, message string) error {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] %s\n", name, message)
			return nil
		},
		RequestValue: func(name, prompt string, secret bool) (string, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] RequestValue called: %s (secret=%v)\n", name, prompt, secret)
			return "", fmt.Errorf("interactive prompts not supported")
		},
		Confirm: func(name, prompt, yes, no string) (bool, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] Confirm called: %s (yes=%s, no=%s)\n", name, prompt, yes, no)
			return false, fmt.Errorf("interactive prompts not supported")
		},
		WaitTimer: func(name string) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] WaitTimer called - plugin is waiting...\n", name)
		},
	}

	pluginRecipient, pluginErr := plugin.NewRecipient(recipientStr, ui)
	if pluginErr == nil {
		return pluginRecipient, nil
	}

	// If both failed, return an appropriate error
	return nil, fmt.Errorf("error parsing recipient file %s as native recipient: %w, as plugin recipient: %v", path, err, pluginErr)
}

// ExtractRecipient extracts a recipient (public key) from any age.Identity,
// whether it's a native X25519 identity or a plugin-based identity.
func ExtractRecipient(identity age.Identity) (age.Recipient, error) {
	switch id := identity.(type) {
	case *age.X25519Identity:
		return id.Recipient(), nil
	case *plugin.Identity:
		return id.Recipient(), nil
	default:
		return nil, fmt.Errorf("unsupported identity type: %T", identity)
	}
}

// ExtractRecipientString extracts the public key string from an identity file.
// This is useful for displaying the public key without needing to convert
// from a Recipient object (which plugin recipients don't support).
// For plugin identities, it looks for a "# public key:" comment in the identity file.
func ExtractRecipientString(identityPath string) (string, error) {
	// Load the identity
	identity, err := LoadIdentity(identityPath)
	if err != nil {
		return "", err
	}

	// Handle based on identity type
	switch id := identity.(type) {
	case *age.X25519Identity:
		// For X25519, we can get the recipient and convert to string
		return id.Recipient().String(), nil
	case *plugin.Identity:
		// For plugin identities, read the file and look for the public key in comments
		// Plugin identity files typically include a "# public key: age1..." line
		content, err := os.ReadFile(identityPath)
		if err != nil {
			return "", fmt.Errorf("failed to read identity file: %w", err)
		}

		// Look for lines starting with "# public key:" or "# recipient:" (case-insensitive)
		lines := bytes.Split(content, []byte("\n"))
		for _, line := range lines {
			trimmed := bytes.TrimSpace(line)
			lowerLine := bytes.ToLower(trimmed)
			if bytes.HasPrefix(lowerLine, []byte("# public key:")) || bytes.HasPrefix(lowerLine, []byte("# recipient:")) {
				// Extract the key after the colon
				parts := bytes.SplitN(trimmed, []byte(":"), 2)
				if len(parts) == 2 {
					return string(bytes.TrimSpace(parts[1])), nil
				}
			}
		}

		// If we couldn't find the recipient in comments, return an error
		return "", fmt.Errorf("could not find public key in identity file; plugin identity files should include a '# public key: age1...' comment line")
	default:
		return "", fmt.Errorf("unsupported identity type: %T", identity)
	}
}

// CopyFile copies a file from source to destination with secure permissions (0600).
// Creates parent directories if needed.
func CopyFile(sourcePath, destPath string) error {
	// Ensure parent directory exists
	if err := config.EnsureParentDir(destPath); err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source file %s: %w", sourcePath, err)
	}
	defer src.Close()

	// Create destination file with secure permissions
	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating destination file %s: %w", destPath, err)
	}
	defer dst.Close()

	// Copy the file
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("error copying file from %s to %s: %w", sourcePath, destPath, err)
	}

	return nil
}

// GetOrCreateVaultKey loads an existing vault key if it exists, or generates a new one if it doesn't.
// This function does NOT save the vault key - that's the caller's responsibility.
func GetOrCreateVaultKey(cfg *config.Config) (age.Identity, error) {
	// Check if vault key file exists
	if _, err := os.Stat(cfg.VaultKeyFile); os.IsNotExist(err) {
		// Vault key doesn't exist, generate a new one
		return vault.GenerateVaultKey()
	} else if err != nil {
		return nil, fmt.Errorf("failed to check vault key file: %w", err)
	}

	// Vault key exists, load and decrypt it
	userIdentity, err := LoadIdentity(cfg.IdentityFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault key file: %w", err)
	}

	vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	// Extract and return the vault key identity
	vaultKeyIdentity, err := vaultKey.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault key identity: %w", err)
	}

	return vaultKeyIdentity, nil
}

// SaveVaultKeyForIdentity encrypts a vault key for a specific identity and saves it to disk.
func SaveVaultKeyForIdentity(vaultKeyIdentity age.Identity, userIdentity age.Identity, savePath string) error {
	// Extract recipient from user identity
	recipient, err := ExtractRecipient(userIdentity)
	if err != nil {
		return fmt.Errorf("failed to extract recipient from identity: %w", err)
	}

	// Encrypt vault key for the recipient
	encryptedVaultKey, err := vault.EncryptVaultKey(vaultKeyIdentity, recipient)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault key: %w", err)
	}

	// Ensure parent directory exists
	if err := config.EnsureParentDir(savePath); err != nil {
		return err
	}

	// Write encrypted vault key to file with secure permissions
	if err := os.WriteFile(savePath, encryptedVaultKey, 0600); err != nil {
		return fmt.Errorf("failed to write vault key file: %w", err)
	}

	return nil
}
