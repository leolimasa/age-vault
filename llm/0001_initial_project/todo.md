# age-vault - Implementation TODO

## Architecture Note
This project is structured as a **library + CLI application**:
- Public library packages (`config`, `vault`, `keymgmt`, `sshagent`) can be imported by other Go programs
- CLI application in `cmd/age-vault/` uses the library packages

## Phase 1: Project Setup & Core Infrastructure

### Directory Structure
- [ ] Create `config/` directory (public library package)
- [ ] Create `vault/` directory (public library package)
- [ ] Create `keymgmt/` directory (public library package)
- [ ] Create `sshagent/` directory (public library package)
- [ ] Create `cmd/age-vault/` directory (CLI application)
- [ ] Create `cmd/age-vault/commands/` directory (CLI command handlers)
- [ ] Delete old `main.go` (will be replaced by `cmd/age-vault/main.go`)

### Library: Configuration Package
- [ ] Implement `config/config.go` (public library package)
  - [ ] Create `Config` struct with all environment variable fields
  - [ ] Implement `NewConfig()` function to read env vars and set defaults
  - [ ] Implement home directory expansion logic
  - [ ] Implement `EnsureConfigDir()` to create config directory
  - [ ] Add config validation
  - [ ] Add package documentation for public API

### Library: Core Vault Operations Package
- [ ] Implement `vault/vault.go` (public library package)
  - [ ] Create `VaultKey` struct wrapper
  - [ ] Implement `GenerateVaultKey()` for creating new vault keys
  - [ ] Implement `EncryptVaultKey()` to encrypt vault key for a recipient
  - [ ] Implement `DecryptVaultKey()` to decrypt vault key with user identity
  - [ ] Implement `VaultKey.Encrypt()` method for encrypting data
  - [ ] Implement `VaultKey.Decrypt()` method for decrypting data
  - [ ] Add package documentation for public API

### Library: Key Management Package
- [ ] Implement `keymgmt/keymgmt.go` (public library package)
  - [ ] Implement `LoadIdentity()` to load age identity from file
  - [ ] Implement `LoadRecipient()` to load age public key from file
  - [ ] Implement `SaveFile()` to securely copy files with proper permissions
  - [ ] Add package documentation for public API

### Phase 1 Testing
- [ ] Write unit tests for `config.NewConfig()` with different env vars
- [ ] Write unit tests for `GenerateVaultKey()`
- [ ] Write unit tests for `EncryptVaultKey()` and `DecryptVaultKey()` cycle
- [ ] Write unit tests for `VaultKey.Encrypt()` and `VaultKey.Decrypt()`
- [ ] Test file loading with `LoadIdentity()` and `LoadRecipient()`
- [ ] Manually test config directory creation
- [ ] validate implementation

## Phase 2: CLI Application - Encrypt/Decrypt Commands

### CLI: Encrypt Command
- [ ] Implement `cmd/age-vault/commands/encrypt.go`
  - [ ] Import library packages: `config`, `vault`, `keymgmt`
  - [ ] Implement `RunEncrypt()` function
  - [ ] Handle file input vs stdin
  - [ ] Handle file output vs stdout
  - [ ] Load user identity using `keymgmt.LoadIdentity()`
  - [ ] Read and decrypt vault key using `vault.DecryptVaultKey()`
  - [ ] Perform encryption operation using `vaultKey.Encrypt()`
  - [ ] Ensure proper cleanup of vault key from memory

### CLI: Decrypt Command
- [ ] Implement `cmd/age-vault/commands/decrypt.go`
  - [ ] Import library packages: `config`, `vault`, `keymgmt`
  - [ ] Implement `RunDecrypt()` function
  - [ ] Handle file input vs stdin
  - [ ] Handle file output vs stdout
  - [ ] Load user identity using `keymgmt.LoadIdentity()`
  - [ ] Read and decrypt vault key using `vault.DecryptVaultKey()`
  - [ ] Perform decryption operation using `vaultKey.Decrypt()`
  - [ ] Ensure proper cleanup of vault key from memory

### CLI: Main Application Setup
- [ ] Implement `cmd/age-vault/main.go`
  - [ ] Add dependency (cobra or use stdlib flag)
  - [ ] Set up CLI framework
  - [ ] Define `encrypt` subcommand with `-o` flag
  - [ ] Define `decrypt` subcommand with `-o` flag
  - [ ] Wire up subcommands to call command handlers
  - [ ] Load config using `config.NewConfig()`
  - [ ] Implement proper error handling and exit codes

### Phase 2 Testing
- [ ] Test encrypting a file: `age-vault encrypt test.txt -o test.enc`
- [ ] Test decrypting a file: `age-vault decrypt test.enc -o test.dec.txt`
- [ ] Test encrypting from stdin: `echo "secret" | age-vault encrypt -o secret.enc`
- [ ] Test decrypting to stdout: `age-vault decrypt secret.enc`
- [ ] Test round-trip: encrypt then decrypt and verify content matches
- [ ] Test error handling for missing vault key
- [ ] Test error handling for missing identity file
- [ ] validate implementation

## Phase 3: CLI Application - Key Management Commands

### CLI: Set Commands
- [ ] Implement `cmd/age-vault/commands/set_vault_key.go`
  - [ ] Import library packages: `config`, `keymgmt`
  - [ ] Implement `RunSetVaultKey()` function
  - [ ] Use `keymgmt.SaveFile()` to copy file with proper permissions (0600)
  - [ ] Create parent directories if needed using `cfg.EnsureConfigDir()`
- [ ] Implement `cmd/age-vault/commands/set_identity.go`
  - [ ] Import library packages: `config`, `keymgmt`
  - [ ] Implement `RunSetIdentity()` function
  - [ ] Use `keymgmt.SaveFile()` to copy identity file with 0600 permissions
- [ ] Implement `cmd/age-vault/commands/set_pubkey.go`
  - [ ] Import library packages: `config`, `keymgmt`
  - [ ] Implement `RunSetPubkey()` function
  - [ ] Use `keymgmt.SaveFile()` to copy public key file with 0644 permissions

### CLI: Encrypt Vault Key Command
- [ ] Implement `cmd/age-vault/commands/encrypt_vault_key.go`
  - [ ] Import library packages: `config`, `vault`, `keymgmt`
  - [ ] Implement `RunEncryptVaultKey()` function
  - [ ] Check if vault key exists, generate using `vault.GenerateVaultKey()` if not
  - [ ] Load existing vault key if present
  - [ ] Load recipient public key using `keymgmt.LoadRecipient()`
  - [ ] Encrypt vault key for recipient using `vault.EncryptVaultKey()`
  - [ ] Write to output file or stdout

### CLI Integration
- [ ] Update `cmd/age-vault/main.go` to add key management subcommands
  - [ ] Define `set-vault-key` subcommand
  - [ ] Define `set-identity` subcommand
  - [ ] Define `set-pubkey` subcommand
  - [ ] Define `encrypt-vault-key` subcommand with `-o` flag
  - [ ] Wire up all subcommands to their handlers

### Phase 3 Testing
- [ ] Test new user workflow:
  - [ ] Generate new age identity: `age-keygen -o test-identity.txt`
  - [ ] Set identity: `age-vault set-identity test-identity.txt`
  - [ ] Set pubkey: `age-vault set-pubkey test-pubkey.txt`
  - [ ] Generate encrypted vault key: `age-vault encrypt-vault-key other-pubkey.txt -o vault-key.age`
  - [ ] Set vault key: `age-vault set-vault-key vault-key.age`
- [ ] Test vault key generation when none exists
- [ ] Test vault key encryption for second user
- [ ] Verify file permissions are set correctly (0600 for sensitive, 0644 for public)
- [ ] validate implementation

## Phase 4: CLI Application - SOPS Integration

### CLI: SOPS Passthrough Command
- [ ] Implement `cmd/age-vault/commands/sops.go`
  - [ ] Import library packages: `config`, `vault`, `keymgmt`
  - [ ] Implement `RunSops()` function
  - [ ] Load user identity using `keymgmt.LoadIdentity()`
  - [ ] Decrypt vault key using `vault.DecryptVaultKey()`
  - [ ] Create temporary file with secure permissions (0600)
  - [ ] Write vault key identity to temp file
  - [ ] Execute sops command with arguments using `exec.Command`
  - [ ] Set up proper environment for sops to find age identity
  - [ ] Ensure temp file cleanup with defer
  - [ ] Capture and return sops exit code

### CLI Integration
- [ ] Update `cmd/age-vault/main.go` to add sops subcommand
  - [ ] Define `sops` subcommand that accepts all remaining args
  - [ ] Wire up to command handler
  - [ ] Handle argument passthrough correctly

### Phase 4 Testing
- [ ] Install sops if needed for testing
- [ ] Create test encrypted file with sops: `sops -e test.yaml > test.enc.yaml`
- [ ] Test decryption: `age-vault sops -d test.enc.yaml`
- [ ] Test encryption: `age-vault sops -e test.yaml`
- [ ] Test in-place edit: `age-vault sops test.enc.yaml`
- [ ] Verify temp file is cleaned up after sops execution
- [ ] Test error handling when sops is not installed
- [ ] validate implementation

## Phase 5: Library + CLI - SSH Agent Implementation

### Library: SSH Agent Package
- [ ] Add SSH dependencies to `go.mod`
  - [ ] Add `golang.org/x/crypto/ssh`
  - [ ] Add `golang.org/x/crypto/ssh/agent`
- [ ] Implement `sshagent/agent.go` (public library package)
  - [ ] Create `VaultSSHAgent` struct
  - [ ] Implement `NewVaultSSHAgent()` to scan keys directory
  - [ ] Implement `List()` to list available SSH keys
  - [ ] Implement `Sign()` to handle signing requests
  - [ ] Implement on-demand decryption of private keys using vault key
  - [ ] Implement `Start()` to create Unix socket and serve agent protocol
  - [ ] Add proper cleanup and shutdown handling
  - [ ] Add package documentation for public API

### CLI: SSH Agent Command
- [ ] Implement `cmd/age-vault/commands/ssh_agent.go`
  - [ ] Import library packages: `config`, `vault`, `keymgmt`, `sshagent`
  - [ ] Implement `RunSSHAgent()` function
  - [ ] Handle keys directory from argument or env var
  - [ ] Load user identity using `keymgmt.LoadIdentity()`
  - [ ] Decrypt vault key using `vault.DecryptVaultKey()`
  - [ ] Create VaultSSHAgent instance using `sshagent.NewVaultSSHAgent()`
  - [ ] Determine socket path (use temp directory)
  - [ ] Start agent using `agent.Start()`
  - [ ] Print environment variables (SSH_AUTH_SOCK, SSH_AGENT_PID)
  - [ ] Handle SIGINT/SIGTERM for graceful shutdown

### CLI Integration
- [ ] Update `cmd/age-vault/main.go` to add ssh-agent subcommand
  - [ ] Define `ssh-agent` subcommand with optional directory argument
  - [ ] Wire up to command handler

### Phase 5 Testing
- [ ] Create test SSH key: `ssh-keygen -t ed25519 -f test-ssh-key`
- [ ] Encrypt SSH key with age-vault: `age-vault encrypt test-ssh-key -o test-ssh-key.age`
- [ ] Set up keys directory with encrypted SSH key
- [ ] Start agent: `age-vault ssh-agent /path/to/keys/dir`
- [ ] Export SSH_AUTH_SOCK variable
- [ ] Test listing keys: `ssh-add -l`
- [ ] Test SSH connection using agent key
- [ ] Test multiple keys in directory
- [ ] Test agent shutdown and cleanup
- [ ] validate implementation

## Phase 6: Final Integration, Documentation & Library Publishing

### Integration Testing
- [ ] Test complete new user onboarding workflow from README
  - [ ] User A creates vault key and encrypts for themselves
  - [ ] User A encrypts some secrets
  - [ ] User B generates new identity
  - [ ] User A encrypts vault key for User B
  - [ ] User B sets vault key and can decrypt secrets
- [ ] Test all CLI commands with environment variable overrides
- [ ] Test with age-plugin-tpm if available (hardware HSM)
- [ ] Test error scenarios:
  - [ ] Missing vault key file
  - [ ] Invalid identity file
  - [ ] Corrupted encrypted data
  - [ ] Wrong permissions on files
  - [ ] Missing environment variables

### Library Testing
- [ ] Create example program that imports and uses the library packages
- [ ] Test importing individual packages: `config`, `vault`, `keymgmt`, `sshagent`
- [ ] Verify library can be used programmatically without CLI
- [ ] Test library integration in a simple Go program

### Documentation
- [ ] Update README.md if needed (already exists, verify accuracy)
- [ ] Add CLI usage examples for each command
- [ ] Add library usage examples showing how to import and use packages
- [ ] Document public API for each package (config, vault, keymgmt, sshagent)
- [ ] Document error messages and troubleshooting
- [ ] Add section on security best practices
- [ ] Document supported age plugins for HSM integration
- [ ] Add example showing how another Go program can use age-vault as a library

### Code Quality
- [ ] Add package-level documentation (doc.go or package comments) for all public packages
- [ ] Add comments to all exported functions, types, and methods
- [ ] Run `go fmt` on all files
- [ ] Run `go vet` and fix any issues
- [ ] Run `golint` or `staticcheck` if available
- [ ] Add proper error context to all error returns
- [ ] Ensure all public APIs have clear, concise documentation

### Build & Release
- [ ] Test CLI build: `go build -o age-vault ./cmd/age-vault`
- [ ] Test CLI installation: `go install ./cmd/age-vault`
- [ ] Verify CLI binary works in clean environment
- [ ] Test that library packages can be imported by external programs
- [ ] Test on different platforms if possible (Linux, macOS)
- [ ] Verify go.mod has correct module path

### Phase 6 Validation
- [ ] Run all tests: `go test ./...`
- [ ] Perform complete workflow test as described in README
- [ ] Verify all commands work as documented
- [ ] Check that vault key never appears on disk in plaintext
- [ ] validate implementation

## Notes

### Architecture
- This project is structured as a **library + CLI application**
- Public packages (`config`, `vault`, `keymgmt`, `sshagent`) are in the root and can be imported by other Go programs
- CLI application is in `cmd/age-vault/` and uses the library packages
- CLI command handlers are thin wrappers that import and call library functions
- This allows the core functionality to be reused programmatically

### Security
- DO NOT commit to git unless explicitly instructed
- Keep security in mind throughout: vault key should only exist decrypted in memory
- Test with small files first before testing with larger files
- Use defer for cleanup operations to ensure they happen even on error paths

### Build Commands
- Build CLI: `go build -o age-vault ./cmd/age-vault`
- Install CLI: `go install ./cmd/age-vault`
- Run tests: `go test ./...`
- Import library: `import "github.com/leolimasa/age-vault/vault"`
