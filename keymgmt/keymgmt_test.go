package keymgmt

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func TestLoadIdentity(t *testing.T) {
	// Generate a test identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("GenerateX25519Identity() failed: %v", err)
	}

	// Create a temp file
	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "identity.txt")

	// Write the identity to the file
	err = os.WriteFile(identityFile, []byte(identity.String()), 0600)
	if err != nil {
		t.Fatalf("Failed to write identity file: %v", err)
	}

	// Load the identity
	loadedIdentity, err := LoadIdentity(identityFile)
	if err != nil {
		t.Fatalf("LoadIdentity() failed: %v", err)
	}

	// Compare the identities by converting to string
	// identity is already *age.X25519Identity from GenerateX25519Identity()
	loadedX25519, ok := loadedIdentity.(*age.X25519Identity)
	if !ok {
		t.Error("Loaded identity is not an X25519Identity")
	}
	if identity.String() != loadedX25519.String() {
		t.Error("Loaded identity does not match original")
	}
}

func TestLoadIdentity_NonExistent(t *testing.T) {
	_, err := LoadIdentity("/nonexistent/identity.txt")
	if err == nil {
		t.Error("LoadIdentity() should fail with non-existent file")
	}
}

func TestLoadRecipient(t *testing.T) {
	// Generate a test identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("GenerateX25519Identity() failed: %v", err)
	}

	// Create a temp file with the public key
	tempDir := t.TempDir()
	recipientFile := filepath.Join(tempDir, "recipient.txt")

	// Write the recipient (public key) to the file
	err = os.WriteFile(recipientFile, []byte(identity.Recipient().String()), 0600)
	if err != nil {
		t.Fatalf("Failed to write recipient file: %v", err)
	}

	// Load the recipient
	loadedRecipient, err := LoadRecipient(recipientFile)
	if err != nil {
		t.Fatalf("LoadRecipient() failed: %v", err)
	}

	// Just verify that the recipient was loaded successfully
	// We can't easily compare recipients directly since age.Recipient is an interface
	if loadedRecipient == nil {
		t.Error("Loaded recipient is nil")
	}
}

func TestLoadRecipient_NonExistent(t *testing.T) {
	_, err := LoadRecipient("/nonexistent/recipient.txt")
	if err == nil {
		t.Error("LoadRecipient() should fail with non-existent file")
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temp source file
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "source.txt")
	testContent := "This is test content"

	err := os.WriteFile(sourceFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	destFile := filepath.Join(tempDir, "subdir", "dest.txt")
	err = CopyFile(sourceFile, destFile)
	if err != nil {
		t.Fatalf("CopyFile() failed: %v", err)
	}

	// Verify the file was copied
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Copied content does not match.\nExpected: %s\nGot: %s", testContent, string(content))
	}

	// Verify permissions
	info, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions to be 0600, got %o", info.Mode().Perm())
	}
}

func TestCopyFile_NonExistentSource(t *testing.T) {
	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "dest.txt")

	err := CopyFile("/nonexistent/source.txt", destFile)
	if err == nil {
		t.Error("CopyFile() should fail with non-existent source")
	}
}
