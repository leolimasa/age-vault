// Package keymgmt provides key management utilities for age-vault.
// It handles loading identities, recipients, and secure file operations.
package keymgmt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/plugin"
	"golang.org/x/term"

	// REVIEW: why are you using fully qualified domain names for local files??
	"github.com/leolimasa/age-vault/config"
	"github.com/leolimasa/age-vault/vault"
)

// NewClientUI creates and returns a configured plugin.ClientUI instance
// that properly prompts the user for input when needed.
func NewClientUI() *plugin.ClientUI {
	return &plugin.ClientUI{
		DisplayMessage: func(name, message string) error {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] %s\n", name, message)
			return nil
		},
		RequestValue: func(name, prompt string, secret bool) (string, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] %s", name, prompt)
			if !strings.HasSuffix(prompt, ": ") && !strings.HasSuffix(prompt, ":") {
				fmt.Fprintf(os.Stderr, ": ")
			}

			if secret {
				// Read password securely without echoing
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintf(os.Stderr, "\n")
				if err != nil {
					return "", fmt.Errorf("failed to read secret value: %w", err)
				}
				return string(passwordBytes), nil
			}

			// Read regular input
			reader := bufio.NewReader(os.Stdin)
			value, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("failed to read value: %w", err)
			}
			return strings.TrimSpace(value), nil
		},
		Confirm: func(name, prompt, yes, no string) (bool, error) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] %s", name, prompt)
			if yes != "" && no != "" {
				fmt.Fprintf(os.Stderr, " [%s/%s]", yes, no)
			}
			fmt.Fprintf(os.Stderr, ": ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return false, fmt.Errorf("failed to read confirmation: %w", err)
			}
			response = strings.ToLower(strings.TrimSpace(response))

			// Check if response matches the "yes" option
			if yes != "" {
				yesLower := strings.ToLower(yes)
				if response == yesLower || response == string(yesLower[0]) {
					return true, nil
				}
			} else {
				// Default yes options
				if response == "y" || response == "yes" {
					return true, nil
				}
			}

			return false, nil
		},
		WaitTimer: func(name string) {
			fmt.Fprintf(os.Stderr, "[PLUGIN %s] Waiting...\n", name)
		},
	}
}

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

	// If native parsing failed, try as plugin identity
	// Extract the identity line (skip comments and empty lines)
	var identityStr string
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		// Skip empty lines and comments
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}
		// Found a non-comment line - assume it's the identity
		identityStr = string(trimmed)
		break
	}

	if identityStr == "" {
		return nil, fmt.Errorf("no identity found in file %s", path)
	}

	// Try loading as plugin identity with proper UI
	ui := NewClientUI()
	pluginIdentity, pluginErr := plugin.NewIdentity(identityStr, ui)
	if pluginErr == nil {
		return pluginIdentity, nil
	}

	// If both failed, return an appropriate error
	return nil, fmt.Errorf("error parsing identity file %s as native identity: %w, as plugin identity: %v", path, err, pluginErr)
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

// RecipientToString converts a recipient to its string representation.
// For X25519Recipients, it uses the String() method.
// For plugin Recipients, we need to read the public key from the identity file comments.
func RecipientToString(recipient age.Recipient) (string, error) {
	switch r := recipient.(type) {
	case *age.X25519Recipient:
		return r.String(), nil
	default:
		// For plugin recipients or other types, we can't easily get the string
		return "", fmt.Errorf("cannot convert recipient type %T to string directly; for plugin recipients, extract the public key from the identity file comments", recipient)
	}
}

// ExtractRecipientString extracts the public key string from an identity.
// For X25519 identities, it gets the recipient and converts to string.
// For plugin identities, it reads the public key from the identity file comments.
func ExtractRecipientString(identityPath string) (string, error) {
	identity, err := LoadIdentity(identityPath)
	if err != nil {
		return "", err
	}

	// For X25519, we can get the recipient and convert to string
	if x25519Identity, ok := identity.(*age.X25519Identity); ok {
		return x25519Identity.Recipient().String(), nil
	}

	// For plugin identities, read the public key from file comments
	if _, ok := identity.(*plugin.Identity); ok {
		content, err := os.ReadFile(identityPath)
		if err != nil {
			return "", fmt.Errorf("failed to read identity file: %w", err)
		}

		// Look for public key in comments
		lines := bytes.Split(content, []byte("\n"))
		for _, line := range lines {
			trimmed := bytes.TrimSpace(line)
			lowerLine := bytes.ToLower(trimmed)
			if bytes.HasPrefix(lowerLine, []byte("# public key:")) || bytes.HasPrefix(lowerLine, []byte("# recipient:")) {
				parts := bytes.SplitN(trimmed, []byte(":"), 2)
				if len(parts) == 2 {
					return string(bytes.TrimSpace(parts[1])), nil
				}
			}
		}
		return "", fmt.Errorf("could not find public key in plugin identity file; plugin identity files should include a '# public key: age1...' comment line")
	}

	return "", fmt.Errorf("unsupported identity type: %T", identity)
}

// VaultKeyFromIdentityFile loads a user identity from file, reads the encrypted
// vault key from disk, decrypts it using the identity, and returns the decrypted
// vault key wrapped in a vault.VaultKey object.
func VaultKeyFromIdentityFile(identityFilePath string, vaultKeyFilePath string) (*vault.VaultKey, error) {
	// Load user's identity
	userIdentity, err := LoadIdentity(identityFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	// Read encrypted vault key
	encryptedVaultKey, err := os.ReadFile(vaultKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault key file: %w", err)
	}

	// Decrypt vault key
	vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, userIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	return vaultKey, nil
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

	// Vault key exists, load and decrypt it using helper function
	vaultKey, err := VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load vault key: %w", err)
	}

	return vaultKey.GetIdentity()
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
