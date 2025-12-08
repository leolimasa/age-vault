# age-vault - Implementation Plan

## Architecture Overview

This project is structured as:
1. **Public library packages** (`config`, `vault`, `keymgmt`, `sshagent`) - can be imported and used by other Go programs
2. **CLI application** (`cmd/age-vault`) - a command-line tool that uses the library

This separation allows the core functionality to be consumed programmatically while providing a user-friendly CLI.

## Data Structures

### `config/config.go`
**Create new file** (Public package)

```go
type Config struct {
    VaultKeyFile     string  // Path to encrypted vault key
    IdentityFile     string  // Path to user's private key (identity)
    SSHKeysDir       string  // Directory containing encrypted SSH keys
}
```
- Stores all configuration paths from environment variables, YAML config file, or defaults
- Supports `age_vault.yml` config file auto-detection by traversing up directory tree
- Provides validation methods to ensure paths are accessible
- **Public package** - can be imported by other programs

### `vault/vault.go`
**Create new file** (Public package)

```go
type VaultKey struct {
    identity age.Identity  // Decrypted vault key as age identity
}
```
- Wrapper for the decrypted vault key that ensures it stays in memory only
- Provides methods for encryption/decryption operations
- **Public package** - can be imported by other programs

## Library Packages (Public API)

### `config/config.go`
**Create new file** (Public package - can be imported by other programs)

#### `NewConfig() (*Config, error)`
- Searches for `age_vault.yml` config file by traversing up directory tree from current working directory
- If found, loads configuration from YAML file (supports: `vault_key_file`, `identity_file`, `ssh_keys_dir`)
- Reads environment variables (override YAML config):
  - `AGE_VAULT_KEY_FILE` (default: `~/.config/.age-vault/vault_key.age`)
  - `AGE_VAULT_IDENTITY_FILE` (default: `~/.config/.age-vault/identity.txt`)
  - `AGE_VAULT_SSH_KEYS_DIR` (no default)
- Expands home directory paths using `os.UserHomeDir()`
- Returns populated Config struct

#### `EnsureParentDir(path string) error`
- Creates parent directory for the given file path if it doesn't exist
- Sets appropriate permissions (0700)
- Used when saving vault keys or identity files to ensure the directory structure exists

### `vault/vault.go`
**Create new file** (Public package - can be imported by other programs)

#### `GenerateVaultKey() (age.Identity, error)`
- Generates a new X25519 age identity to serve as the vault key
- Returns the identity (private key) that will be encrypted per-user
- Used when initializing a new vault

#### `EncryptVaultKey(vaultKey age.Identity, recipientPubKey age.Recipient) ([]byte, error)`
- Takes a vault key identity and a user's public key (recipient)
- Encrypts the vault key's private key string using age encryption
- Returns the encrypted bytes
- Used when adding new users to the vault

#### `DecryptVaultKey(encryptedKey []byte, userIdentity age.Identity) (*VaultKey, error)`
- Takes encrypted vault key bytes and user's identity (private key)
- Decrypts the vault key using the user's private key
- Parses the decrypted content as an age identity
- Returns VaultKey wrapper containing the decrypted vault key
- This is the core function that unwraps the vault key for use

#### `(vk *VaultKey) Encrypt(input io.Reader, output io.Writer) error`
- Takes input stream and output stream
- Uses the vault key to create an age encryptor
- Encrypts data from input to output using age.Encrypt
- Returns any encryption errors

#### `(vk *VaultKey) Decrypt(input io.Reader, output io.Writer) error`
- Takes input stream and output stream
- Uses the vault key to create an age decryptor
- Decrypts data from input to output using age.Decrypt
- Returns any decryption errors

### `keymgmt/keymgmt.go`
**Create new file** (Public package - can be imported by other programs)

#### `LoadIdentity(path string) (age.Identity, error)`
- Reads an age identity file from the given path
- Parses it using age.ParseIdentities
- Returns the first identity found
- Used to load user's private key

#### `LoadRecipient(path string) (age.Recipient, error)`
- Reads an age public key file from the given path
- Parses it using age.ParseRecipients
- Returns the first recipient found
- Used to load user's public key

#### `CopyFile(sourcePath, destPath string) error`
- Copies a file from source to destination
- Used for `vault-key set` and `identity set` commands
- Creates parent directories if needed using `config.EnsureParentDir()`
- Sets appropriate file permissions (0600 for sensitive files)

### `sshagent/agent.go`
**Create new file** (Public package - can be imported by other programs)

#### `NewVaultSSHAgent(keysDir string, vaultKey *vault.VaultKey) (*VaultSSHAgent, error)`
- Creates a new SSH agent implementation
- Scans the keys directory for `.age` encrypted files
- Stores references to encrypted key files and the vault key
- Returns agent instance that implements the ssh-agent protocol

#### `(a *VaultSSHAgent) List() ([]*agent.Key, error)`
- Lists all available SSH keys (returns public key info)
- Decrypts each key file to extract the public key component
- Returns list without actually loading private keys into memory yet

#### `(a *VaultSSHAgent) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error)`
- Handles signing requests from SSH client
- Finds the matching encrypted private key file
- Decrypts the private key using vault key
- Performs the signature operation in memory
- Clears the decrypted private key from memory
- Returns signature

#### `(a *VaultSSHAgent) Start(socketPath string) error`
- Creates a Unix socket at the specified path
- Listens for SSH agent protocol connections
- Handles agent requests (list, sign, etc.)
- Runs until interrupted

## CLI Application

### `cmd/age-vault/main.go`
**Create new file**

#### `main()`
- Sets up CLI framework (using `cobra` or `flag` package)
- Defines hierarchical command structure:
  - Top-level commands: `encrypt`, `decrypt`, `sops`
  - `vault-key` command group with subcommands: `encrypt`, `set`
  - `identity` command group with subcommands: `set`, `pubkey`
  - `ssh` command group with subcommands: `start-agent`, `list-keys`
- Each subcommand delegates to corresponding handler in `cmd/age-vault/commands/`
- Loads config once at startup using `config.NewConfig()`
- Handles global error reporting and exit codes
- This is the CLI entry point that imports and uses the library packages

### `cmd/age-vault/commands/encrypt.go`
**Create new file** (CLI-specific command handler)

#### `RunEncrypt(inputPath, outputPath string, cfg *config.Config) error`
- Imports library packages: `config`, `vault`, `keymgmt`
- Loads user's identity from config using `keymgmt.LoadIdentity()`
- Reads encrypted vault key file
- Decrypts the vault key using `vault.DecryptVaultKey()`
- Opens input (file or stdin)
- Opens output (file or stdout)
- Calls `vaultKey.Encrypt(input, output)`
- Ensures vault key is cleared from memory after use

### `cmd/age-vault/commands/decrypt.go`
**Create new file** (CLI-specific command handler)

#### `RunDecrypt(inputPath, outputPath string, cfg *config.Config) error`
- Imports library packages: `config`, `vault`, `keymgmt`
- Loads user's identity from config using `keymgmt.LoadIdentity()`
- Reads encrypted vault key file
- Decrypts the vault key using `vault.DecryptVaultKey()`
- Opens input (file or stdin)
- Opens output (file or stdout)
- Calls `vaultKey.Decrypt(input, output)`
- Ensures vault key is cleared from memory after use

### `cmd/age-vault/commands/sops.go`
**Create new file** (CLI-specific command handler)

#### `RunSops(sopsArgs []string, cfg *config.Config) error`
- Imports library packages: `config`, `vault`, `keymgmt`
- Loads user's identity and reads vault key file
- Decrypts the vault key using `vault.DecryptVaultKey()`
- Creates a pipe (file descriptor) containing the vault key identity
- Uses process substitution to pass the file descriptor to sops via `/dev/fd/N` on Unix
- Sets `SOPS_AGE_KEY_FILE` environment variable to point to the file descriptor
- Executes `sops` command with provided arguments using `exec.Command`
- Vault key never touches disk, only exists in memory and the pipe buffer
- Returns sops exit code

### `cmd/age-vault/commands/ssh_start_agent.go`
**Create new file** (CLI-specific command handler)

#### `RunSSHStartAgent(keysDir string, cfg *config.Config) error`
- Imports library packages: `config`, `vault`, `keymgmt`, `sshagent`
- Uses keysDir from argument or falls back to `cfg.SSHKeysDir`
- Loads user's identity and reads vault key file
- Decrypts vault key using `vault.DecryptVaultKey()`
- Creates VaultSSHAgent instance using `sshagent.NewVaultSSHAgent()`
- Determines socket path (creates in temp directory)
- Starts the agent with `agent.Start(socketPath)`
- Prints environment variables for user to set: `SSH_AUTH_SOCK` and `SSH_AGENT_PID`
- Handles shutdown signals gracefully (SIGINT/SIGTERM)

### `cmd/age-vault/commands/ssh_list_keys.go`
**Create new file** (CLI-specific command handler)

#### `RunSSHListKeys(cfg *config.Config) error`
- Imports library packages: `config`
- Reads the SSH keys directory from `cfg.SSHKeysDir`
- Lists all `.age` encrypted files in the directory
- Displays the key file names to the user
- Does not decrypt keys, only shows what's available

### `cmd/age-vault/commands/vault_key_encrypt.go`
**Create new file** (CLI-specific command handler)

#### `RunVaultKeyEncrypt(pubKeyPath, outputPath string, cfg *config.Config) error`
- Imports library packages: `config`, `vault`, `keymgmt`
- Checks if vault key file exists at `cfg.VaultKeyFile`
- If not exists: generates new vault key using `vault.GenerateVaultKey()` and saves encrypted version
- If exists: loads user identity and decrypts existing vault key
- Loads the recipient's public key from `pubKeyPath` using `keymgmt.LoadRecipient()`
- Encrypts vault key for recipient using `vault.EncryptVaultKey()`
- Writes encrypted output to file or stdout
- Returns encrypted vault key that can be sent to new user

### `cmd/age-vault/commands/vault_key_set.go`
**Create new file** (CLI-specific command handler)

#### `RunVaultKeySet(sourcePath string, cfg *config.Config) error`
- Imports library packages: `config`, `keymgmt`
- Copies file from `sourcePath` to `cfg.VaultKeyFile` using `keymgmt.CopyFile()`
- Creates parent directory if needed
- Sets permissions to 0600

### `cmd/age-vault/commands/identity_set.go`
**Create new file** (CLI-specific command handler)

#### `RunIdentitySet(sourcePath string, cfg *config.Config) error`
- Imports library packages: `config`, `keymgmt`
- Copies identity file from `sourcePath` to `cfg.IdentityFile` using `keymgmt.CopyFile()`
- Creates parent directory if needed
- Sets permissions to 0600

### `cmd/age-vault/commands/identity_pubkey.go`
**Create new file** (CLI-specific command handler)

#### `RunIdentityPubkey(outputPath string, cfg *config.Config) error`
- Imports library packages: `config`, `keymgmt`
- Loads user identity from `cfg.IdentityFile` using `keymgmt.LoadIdentity()`
- Extracts public key from the identity
- Writes public key to output file or stdout
- Used to share public key with vault administrators for encryption

## Key Implementation Notes

### Security Considerations
1. **Memory Management**: Use `runtime.KeepAlive()` and explicit zeroing for sensitive data where possible
2. **File Permissions**: All identity and vault key files use 0600 permissions
3. **Process Substitution**: For sops integration, use file descriptor pipes to pass vault key without writing to disk
4. **No Disk Persistence**: Vault key never written to disk in plaintext form, even in temporary files

### Error Handling
- All functions return errors following Go conventions
- File operations check for existence and permissions
- Provide helpful error messages indicating which file or operation failed

### Testing Strategy
- Unit tests for each vault operation (encrypt/decrypt)
- Integration tests for full encrypt/decrypt cycle
- Mock file system operations for testing
- Test environment variable handling with different configurations

### Dependencies
- `filippo.io/age` - already in go.mod
- `golang.org/x/crypto/ssh` - for SSH agent implementation
- `golang.org/x/crypto/ssh/agent` - for SSH agent protocol
- CLI framework: consider `cobra` for better subcommand handling, or use stdlib `flag` for simplicity

### Directory Structure
```
age-vault/
├── config/                           (public library package)
│   └── config.go                     (create)
├── vault/                            (public library package)
│   └── vault.go                      (create)
├── keymgmt/                          (public library package)
│   └── keymgmt.go                    (create)
├── sshagent/                         (public library package)
│   └── agent.go                      (create)
├── cmd/
│   └── age-vault/                    (CLI application)
│       ├── main.go                   (create)
│       └── commands/                 (CLI command handlers)
│           ├── encrypt.go            (create)
│           ├── decrypt.go            (create)
│           ├── sops.go               (create)
│           ├── ssh_start_agent.go    (create)
│           ├── ssh_list_keys.go      (create)
│           ├── vault_key_encrypt.go  (create)
│           ├── vault_key_set.go      (create)
│           ├── identity_set.go       (create)
│           └── identity_pubkey.go    (create)
├── main.go                           (delete - replaced by cmd/age-vault/main.go)
├── go.mod                            (modify - update module path if needed)
├── age_vault.yml                     (optional - user creates for per-project config)
└── README.md                         (existing)
```

### Library Usage Example

Other Go programs can import and use the library:

```go
package main

import (
    "os"
    "github.com/leolimasa/age-vault/config"
    "github.com/leolimasa/age-vault/vault"
    "github.com/leolimasa/age-vault/keymgmt"
)

func main() {
    // Load configuration
    cfg, err := config.NewConfig()
    if err != nil {
        panic(err)
    }

    // Load user identity
    identity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
    if err != nil {
        panic(err)
    }

    // Read encrypted vault key
    encryptedVaultKey, err := os.ReadFile(cfg.VaultKeyFile)
    if err != nil {
        panic(err)
    }

    // Decrypt vault key
    vaultKey, err := vault.DecryptVaultKey(encryptedVaultKey, identity)
    if err != nil {
        panic(err)
    }

    // Use vault key to encrypt data
    input, _ := os.Open("plaintext.txt")
    output, _ := os.Create("encrypted.age")
    defer input.Close()
    defer output.Close()

    err = vaultKey.Encrypt(input, output)
    if err != nil {
        panic(err)
    }
}
```
