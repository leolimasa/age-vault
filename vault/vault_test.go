package vault

import (
	"bytes"
	"strings"
	"testing"

	"filippo.io/age"
)

func TestGenerateVaultKey(t *testing.T) {
	vaultKey, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey() failed: %v", err)
	}

	if vaultKey == nil {
		t.Error("GenerateVaultKey() returned nil")
	}

	// Check that the vault key can be converted to a string
	x25519Identity, ok := vaultKey.(*age.X25519Identity)
	if !ok {
		t.Error("Generated vault key is not an X25519Identity")
	}
	keyStr := x25519Identity.String()
	if keyStr == "" {
		t.Error("Generated vault key has empty string representation")
	}

	// Check that it starts with the expected prefix
	if !strings.HasPrefix(keyStr, "AGE-SECRET-KEY-") {
		t.Errorf("Generated vault key has unexpected format: %s", keyStr)
	}
}

func TestEncryptDecryptVaultKey(t *testing.T) {
	// Generate a vault key
	vaultKey, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey() failed: %v", err)
	}

	// Generate a user identity
	userIdentity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("GenerateX25519Identity() failed: %v", err)
	}

	// Encrypt the vault key for the user
	encryptedKey, err := EncryptVaultKey(vaultKey, userIdentity.Recipient())
	if err != nil {
		t.Fatalf("EncryptVaultKey() failed: %v", err)
	}

	if len(encryptedKey) == 0 {
		t.Error("EncryptVaultKey() returned empty bytes")
	}

	// Decrypt the vault key
	decryptedVaultKey, err := DecryptVaultKey(encryptedKey, userIdentity)
	if err != nil {
		t.Fatalf("DecryptVaultKey() failed: %v", err)
	}

	// Verify that the decrypted vault key matches the original
	origX25519, _ := vaultKey.(*age.X25519Identity)
	decryptedX25519, ok := decryptedVaultKey.identity.(*age.X25519Identity)
	if !ok {
		t.Error("Decrypted vault key is not an X25519Identity")
	}
	if origX25519.String() != decryptedX25519.String() {
		t.Error("Decrypted vault key does not match original")
	}
}

func TestVaultKeyEncryptDecrypt(t *testing.T) {
	// Generate a vault key
	vaultKey, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey() failed: %v", err)
	}

	// Create a VaultKey wrapper
	vk := &VaultKey{identity: vaultKey}

	// Test data
	plaintext := "This is a secret message!"

	// Encrypt the data
	var encryptedBuf bytes.Buffer
	err = vk.Encrypt(strings.NewReader(plaintext), &encryptedBuf)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	if encryptedBuf.Len() == 0 {
		t.Error("Encrypt() produced no output")
	}

	// Decrypt the data
	var decryptedBuf bytes.Buffer
	err = vk.Decrypt(&encryptedBuf, &decryptedBuf)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	// Verify the decrypted data matches the original
	if decryptedBuf.String() != plaintext {
		t.Errorf("Decrypted data does not match original.\nExpected: %s\nGot: %s", plaintext, decryptedBuf.String())
	}
}

func TestVaultKeyEncryptDecryptLargeData(t *testing.T) {
	// Generate a vault key
	vaultKey, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey() failed: %v", err)
	}

	// Create a VaultKey wrapper
	vk := &VaultKey{identity: vaultKey}

	// Create large test data (1 MB)
	plaintext := strings.Repeat("This is a secret message! ", 40000)

	// Encrypt the data
	var encryptedBuf bytes.Buffer
	err = vk.Encrypt(strings.NewReader(plaintext), &encryptedBuf)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	// Decrypt the data
	var decryptedBuf bytes.Buffer
	err = vk.Decrypt(&encryptedBuf, &decryptedBuf)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	// Verify the decrypted data matches the original
	if decryptedBuf.String() != plaintext {
		t.Errorf("Decrypted large data does not match original. Size: expected %d, got %d", len(plaintext), decryptedBuf.Len())
	}
}

func TestDecryptVaultKeyInvalidData(t *testing.T) {
	// Generate a user identity
	userIdentity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("GenerateX25519Identity() failed: %v", err)
	}

	// Try to decrypt invalid data
	invalidData := []byte("this is not encrypted data")
	_, err = DecryptVaultKey(invalidData, userIdentity)
	if err == nil {
		t.Error("DecryptVaultKey() should fail with invalid data")
	}
}

func TestVaultKeyDecryptInvalidData(t *testing.T) {
	// Generate a vault key
	vaultKey, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey() failed: %v", err)
	}

	// Create a VaultKey wrapper
	vk := &VaultKey{identity: vaultKey}

	// Try to decrypt invalid data
	invalidData := bytes.NewReader([]byte("this is not encrypted data"))
	var output bytes.Buffer
	err = vk.Decrypt(invalidData, &output)
	if err == nil {
		t.Error("Decrypt() should fail with invalid data")
	}
}
