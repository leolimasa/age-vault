// Package keymgmt provides key management utilities for age-vault.
// It handles loading identities, recipients, and secure file operations.
package keymgmt

import (
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"github.com/leolimasa/age-vault/config"
)

// LoadIdentity reads and parses an age identity file from the given path.
// Returns the first identity found in the file.
func LoadIdentity(path string) (age.Identity, error) {
	// Open and parse the identity file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening identity file %s: %w", path, err)
	}
	defer f.Close()

	identities, err := age.ParseIdentities(f)
	if err != nil {
		return nil, fmt.Errorf("error parsing identity file %s: %w", path, err)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities found in file %s", path)
	}

	return identities[0], nil
}

// LoadRecipient reads and parses an age public key file from the given path.
// Returns the first recipient found in the file.
func LoadRecipient(path string) (age.Recipient, error) {
	// Open and parse the recipient file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening recipient file %s: %w", path, err)
	}
	defer f.Close()

	recipients, err := age.ParseRecipients(f)
	if err != nil {
		return nil, fmt.Errorf("error parsing recipient file %s: %w", path, err)
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("no recipients found in file %s", path)
	}

	return recipients[0], nil
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
