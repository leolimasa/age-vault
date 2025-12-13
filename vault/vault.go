// Package vault provides core vault key operations for age-vault.
// It manages vault key generation, encryption, decryption, and data operations.
package vault

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
)

// VaultKey wraps the decrypted vault key and ensures it stays in memory only.
type VaultKey struct {
	identity age.Identity
}

// GenerateVaultKey generates a new X25519 age identity to serve as the vault key.
// This vault key will be encrypted per-user and used to encrypt/decrypt secrets.
func GenerateVaultKey() (age.Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("error generating vault key: %w", err)
	}
	return identity, nil
}

// EncryptVaultKey encrypts a vault key identity for a specific recipient (user's public key).
// Returns the encrypted vault key bytes that can be stored and distributed to users.
func EncryptVaultKey(vaultKey age.Identity, recipientPubKey age.Recipient) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Starting encryption\n")
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Recipient type: %T\n", recipientPubKey)

	// Convert the vault key identity to its string representation
	// X25519Identity has a String() method
	x25519Identity, ok := vaultKey.(*age.X25519Identity)
	if !ok {
		return nil, fmt.Errorf("vault key must be an X25519Identity")
	}
	vaultKeyStr := x25519Identity.String()
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Vault key converted to string\n")

	// Create a buffer to hold the encrypted vault key
	var encryptedBuf bytes.Buffer

	// Create an encryptor for the recipient
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Creating encryptor...\n")
	w, err := age.Encrypt(&encryptedBuf, recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("error creating encryptor: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Encryptor created\n")

	// Write the vault key to the encryptor
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Writing vault key to encryptor...\n")
	if _, err := io.WriteString(w, vaultKeyStr); err != nil {
		return nil, fmt.Errorf("error writing vault key: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Vault key written\n")

	// Close the encryptor to finalize encryption
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Closing encryptor...\n")
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error closing encryptor: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG vault.EncryptVaultKey] Encryptor closed successfully\n")

	return encryptedBuf.Bytes(), nil
}

// DecryptVaultKey decrypts an encrypted vault key using the user's identity (private key).
// Supports both native X25519 identities and plugin-based identities.
// Returns a VaultKey wrapper containing the decrypted vault key ready for use.
func DecryptVaultKey(encryptedKey []byte, userIdentity age.Identity) (*VaultKey, error) {
	// Create a reader for the encrypted vault key
	r := bytes.NewReader(encryptedKey)

	// Decrypt the vault key
	decryptor, err := age.Decrypt(r, userIdentity)
	if err != nil {
		return nil, fmt.Errorf("error creating decryptor: %w", err)
	}

	// Read the decrypted vault key
	var decryptedBuf bytes.Buffer
	if _, err := io.Copy(&decryptedBuf, decryptor); err != nil {
		return nil, fmt.Errorf("error reading decrypted vault key: %w", err)
	}

	// Parse the decrypted vault key as an age identity
	identities, err := age.ParseIdentities(bytes.NewReader(decryptedBuf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("error parsing vault key identity: %w", err)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities found in decrypted vault key")
	}

	return &VaultKey{identity: identities[0]}, nil
}

// Encrypt encrypts data from input to output using the vault key.
func (vk *VaultKey) Encrypt(input io.Reader, output io.Writer) error {
	// Get the recipient from the vault key identity
	x25519Identity, ok := vk.identity.(*age.X25519Identity)
	if !ok {
		return fmt.Errorf("vault key must be an X25519Identity")
	}
	recipient := x25519Identity.Recipient()

	// Create an encryptor
	w, err := age.Encrypt(output, recipient)
	if err != nil {
		return fmt.Errorf("error creating encryptor: %w", err)
	}

	// Copy data from input to encryptor
	if _, err := io.Copy(w, input); err != nil {
		return fmt.Errorf("error encrypting data: %w", err)
	}

	// Close the encryptor to finalize encryption
	if err := w.Close(); err != nil {
		return fmt.Errorf("error closing encryptor: %w", err)
	}

	return nil
}

// Decrypt decrypts data from input to output using the vault key.
func (vk *VaultKey) Decrypt(input io.Reader, output io.Writer) error {
	// Create a decryptor
	r, err := age.Decrypt(input, vk.identity)
	if err != nil {
		return fmt.Errorf("error creating decryptor: %w", err)
	}

	// Copy decrypted data to output
	if _, err := io.Copy(output, r); err != nil {
		return fmt.Errorf("error decrypting data: %w", err)
	}

	return nil
}

// GetIdentity returns the underlying age identity from the vault key.
// This is used when re-encrypting the vault key for new recipients.
func (vk *VaultKey) GetIdentity() (age.Identity, error) {
	if vk.identity == nil {
		return nil, fmt.Errorf("vault key identity is nil")
	}
	return vk.identity, nil
}
