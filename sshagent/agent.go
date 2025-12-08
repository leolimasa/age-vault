// Package sshagent provides SSH agent functionality for age-vault.
// It manages vault-encrypted SSH keys and provides on-demand decryption.
package sshagent

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"filippo.io/age"
	"github.com/leolimasa/age-vault/vault"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// VaultSSHAgent implements an SSH agent that decrypts keys using the vault key.
type VaultSSHAgent struct {
	keysDir  string
	vaultKey *vault.VaultKey
	keys     []sshKeyInfo
}

type sshKeyInfo struct {
	path      string
	publicKey ssh.PublicKey
	signer    ssh.Signer
}

// NewVaultSSHAgent creates a new SSH agent for vault-encrypted keys.
func NewVaultSSHAgent(keysDir string, vaultKey *vault.VaultKey) (*VaultSSHAgent, error) {
	fmt.Fprintf(os.Stderr, "Loading SSH keys from directory: %s\n", keysDir)

	// Scan the keys directory for .age files
	matches, err := filepath.Glob(filepath.Join(keysDir, "*.age"))
	if err != nil {
		return nil, fmt.Errorf("error scanning keys directory: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Found %d encrypted SSH keys\n", len(matches))

	a := &VaultSSHAgent{
		keysDir:  keysDir,
		vaultKey: vaultKey,
		keys:     make([]sshKeyInfo, 0, len(matches)),
	}

	// Load each key
	for _, keyPath := range matches {
		if err := a.loadKey(keyPath); err != nil {
			// Skip keys that fail to load
			fmt.Fprintf(os.Stderr, "Warning: failed to load key %s: %v\n", keyPath, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "Loaded key: %s\n", filepath.Base(keyPath))
	}

	return a, nil
}

// loadKey loads and decrypts an SSH key from the vault.
func (a *VaultSSHAgent) loadKey(keyPath string) error {
	// Read encrypted key file
	encryptedKey, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("error reading key file: %w", err)
	}

	// Decrypt the key using the vault key
	vaultKeyIdentity, err := a.vaultKey.GetIdentity()
	if err != nil {
		return fmt.Errorf("error getting vault key identity: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(encryptedKey), vaultKeyIdentity)
	if err != nil {
		return fmt.Errorf("error decrypting key: %w", err)
	}

	var decryptedKey bytes.Buffer
	if _, err := io.Copy(&decryptedKey, r); err != nil {
		return fmt.Errorf("error reading decrypted key: %w", err)
	}

	// Parse the SSH private key
	signer, err := ssh.ParsePrivateKey(decryptedKey.Bytes())
	if err != nil {
		return fmt.Errorf("error parsing SSH private key: %w", err)
	}

	a.keys = append(a.keys, sshKeyInfo{
		path:      keyPath,
		publicKey: signer.PublicKey(),
		signer:    signer,
	})

	return nil
}

// List returns the list of available SSH keys.
func (a *VaultSSHAgent) List() ([]*agent.Key, error) {
	keys := make([]*agent.Key, len(a.keys))
	for i, k := range a.keys {
		keys[i] = &agent.Key{
			Format:  k.publicKey.Type(),
			Blob:    k.publicKey.Marshal(),
			Comment: filepath.Base(k.path),
		}
	}
	return keys, nil
}

// Sign signs data with the specified public key.
func (a *VaultSSHAgent) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	// Find the key
	for _, k := range a.keys {
		if keysEqual(k.publicKey, key) {
			fmt.Fprintf(os.Stderr, "Signing request with key: %s\n", filepath.Base(k.path))
			return k.signer.Sign(nil, data)
		}
	}
	fmt.Fprintf(os.Stderr, "Sign error: key not found\n")
	return nil, fmt.Errorf("key not found")
}

// SignWithFlags signs data with the specified public key and flags (required by ExtendedAgent).
func (a *VaultSSHAgent) SignWithFlags(key ssh.PublicKey, data []byte, flags agent.SignatureFlags) (*ssh.Signature, error) {
	// Find the key
	for _, k := range a.keys {
		if keysEqual(k.publicKey, key) {
			fmt.Fprintf(os.Stderr, "Signing request (with flags) with key: %s\n", filepath.Base(k.path))
			if signer, ok := k.signer.(ssh.AlgorithmSigner); ok {
				var algo string
				switch flags {
				case agent.SignatureFlagRsaSha256:
					algo = ssh.KeyAlgoRSASHA256
				case agent.SignatureFlagRsaSha512:
					algo = ssh.KeyAlgoRSASHA512
				default:
					algo = ""
				}
				if algo != "" {
					return signer.SignWithAlgorithm(nil, data, algo)
				}
			}
			return k.signer.Sign(nil, data)
		}
	}
	fmt.Fprintf(os.Stderr, "SignWithFlags error: key not found\n")
	return nil, fmt.Errorf("key not found")
}

// keysEqual compares two SSH public keys for equality.
func keysEqual(a, b ssh.PublicKey) bool {
	return bytes.Equal(a.Marshal(), b.Marshal())
}

// reloadKeys reloads all keys from the vault directory.
// This ensures new keys added since the agent started are picked up.
func (a *VaultSSHAgent) reloadKeys() error {
	fmt.Fprintf(os.Stderr, "Reloading keys from vault directory\n")

	// Clear existing keys
	a.keys = make([]sshKeyInfo, 0)

	// Scan the keys directory for .age files
	matches, err := filepath.Glob(filepath.Join(a.keysDir, "*.age"))
	if err != nil {
		return fmt.Errorf("error scanning keys directory: %w", err)
	}

	// Load each key
	for _, keyPath := range matches {
		if err := a.loadKey(keyPath); err != nil {
			// Skip keys that fail to load
			fmt.Fprintf(os.Stderr, "Warning: failed to reload key %s: %v\n", keyPath, err)
			continue
		}
	}

	return nil
}

// Signers returns all signers (required by agent.Agent interface).
func (a *VaultSSHAgent) Signers() ([]ssh.Signer, error) {
	signers := make([]ssh.Signer, len(a.keys))
	for i, k := range a.keys {
		signers[i] = k.signer
	}
	return signers, nil
}

// Start starts the SSH agent on a Unix socket.
func (a *VaultSSHAgent) Start(socketPath string) error {
	// Remove old socket if it exists
	os.Remove(socketPath)

	// Create Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	// Set socket permissions
	if err := os.Chmod(socketPath, 0600); err != nil {
		return fmt.Errorf("error setting socket permissions: %w", err)
	}

	fmt.Printf("SSH agent started on %s\n", socketPath)
	fmt.Printf("export SSH_AUTH_SOCK=%s\n", socketPath)
	fmt.Printf("export SSH_AGENT_PID=%d\n", os.Getpid())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start accepting connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				// Reload keys before serving each connection
				if err := a.reloadKeys(); err != nil {
					fmt.Fprintf(os.Stderr, "Error reloading keys: %v\n", err)
				}
				agent.ServeAgent(a, c)
			}(conn)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down SSH agent...")

	return nil
}

// Additional methods required by agent.Agent interface

func (a *VaultSSHAgent) Add(key agent.AddedKey) error {
	return fmt.Errorf("adding keys not supported")
}

func (a *VaultSSHAgent) Remove(key ssh.PublicKey) error {
	return fmt.Errorf("removing keys not supported")
}

func (a *VaultSSHAgent) RemoveAll() error {
	return fmt.Errorf("removing all keys not supported")
}

func (a *VaultSSHAgent) Lock(passphrase []byte) error {
	return fmt.Errorf("locking not supported")
}

func (a *VaultSSHAgent) Unlock(passphrase []byte) error {
	return fmt.Errorf("unlocking not supported")
}

func (a *VaultSSHAgent) Extension(extensionType string, contents []byte) ([]byte, error) {
	return nil, agent.ErrExtensionUnsupported
}

// Unused interface methods for compatibility
var _ agent.Agent = (*VaultSSHAgent)(nil)
var _ agent.ExtendedAgent = (*VaultSSHAgent)(nil)
