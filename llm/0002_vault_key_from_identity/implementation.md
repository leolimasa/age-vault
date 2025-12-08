# Implementation Plan

## Data Structures

### Modified

**config/config.go - Config struct**
- Add `configFileDir` field (string) to track the directory containing the loaded age_vault.yml file
- This will be used to resolve relative paths from the config file location

## Functions/Methods to Implement

### cmd/age-vault/main.go

**Modified: main()**
- Add new subcommand `vault-key from-identity` under the `vaultKeyCmd` command group
- Wire up the command to call `commands.RunVaultKeyFromIdentity(cfg)`

### cmd/age-vault/commands/vault_key_from_identity.go

**Created: RunVaultKeyFromIdentity(cfg *config.Config) error**
- Check if vault key file already exists at `cfg.VaultKeyFile`
  - If exists, return error "vault key already exists"
- Load the user's identity from `cfg.IdentityFile` using `keymgmt.LoadIdentity()`
- Generate a new vault key using `vault.GenerateVaultKey()`
- Extract the recipient (public key) from the loaded identity
  - Cast identity to `*age.X25519Identity`
  - Call `.Recipient()` method
- Encrypt the vault key for this recipient using `vault.EncryptVaultKey()`
- Ensure parent directory exists for `cfg.VaultKeyFile` using `config.EnsureParentDir()`
- Write encrypted vault key to `cfg.VaultKeyFile` with permissions 0600
- Print success message to stdout

### config/config.go

**Modified: Config struct**
- Add private field `configFileDir string` to store the directory of the loaded config file

**Modified: NewConfig() (*Config, error)**
- Update to pass the config file directory to the path resolution logic
- Call new helper function `resolveConfigPath()` for each path field when loading from YAML

**Modified: findAndLoadYAMLConfig() (yamlConfig, string, error)**
- Change return type to include the directory path where the config file was found
- Return the directory path as second return value

**Created: resolveConfigPath(path string, configFileDir string) string**
- If path is empty, return empty string
- If path is absolute, return as-is
- If path starts with `~`, expand home directory using existing `expandHomePath()`
- Otherwise, treat as relative path and join with `configFileDir`
- Return the resolved absolute path

### sshagent/agent.go

**Modified: NewVaultSSHAgent(keysDir string, vaultKey *vault.VaultKey) (*VaultSSHAgent, error)**
- Add logging: print "Loading SSH keys from directory: <keysDir>" to stderr
- When scanning for .age files, print count: "Found N encrypted SSH keys" to stderr
- In the key loading loop, add logging for each successful load: "Loaded key: <filename>" to stderr
- The existing warning for failed keys already provides error logging

**Modified: loadKey(keyPath string) error**
- No changes needed - this is called by NewVaultSSHAgent which will add the logging

**Modified: Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error)**
- Before returning signature, add logging: "Signing request with key: <comment/path>" to stderr
- If key not found, log error before returning: "Sign error: key not found" to stderr

**Modified: SignWithFlags(key ssh.PublicKey, data []byte, flags agent.SignatureFlags) (*ssh.Signature, error)**
- Before returning signature, add logging: "Signing request (with flags) with key: <comment/path>" to stderr
- If key not found, log error before returning: "SignWithFlags error: key not found" to stderr

**Modified: Start(socketPath string) error**
- In the connection handling goroutine, reload keys before each connection
- Create new method `reloadKeys()` and call it before `agent.ServeAgent()`
- Add logging when reloading: "Reloading keys from vault directory" to stderr

**Created: reloadKeys() error**
- Clear the existing `a.keys` slice
- Re-scan the keys directory for .age files using `filepath.Glob()`
- For each key file, call `loadKey()` to decrypt and load it
- This ensures new keys added to the directory are picked up
- Return any errors encountered during reload

### test/start_ssh_server.sh

**Created: Shell script to start a temporary SSH daemon**
- Accept public key file path as first argument
- Validate that argument is provided
- Create temporary directory for SSH server configuration
- Generate temporary SSH host keys (RSA, ED25519)
- Create temporary sshd_config file with:
  - Listen on localhost only (127.0.0.1)
  - Use random available port (or configurable port)
  - Set `AuthorizedKeysFile` to point to temporary file containing the provided public key
  - Disable password authentication
  - Enable public key authentication
  - Set minimal logging for debugging
  - Use the temporary directory for pid file
- Copy provided public key to authorized_keys location
- Start sshd in foreground mode with the temporary config
- On script exit (trap), clean up temporary directory and stop sshd
- Print connection information (hostname, port, username) for testing

### test/run_integration_tests.sh

**Created: Shell script for integration testing**
- Build the age-vault binary using `go build -o ./test/age-vault ./cmd/age-vault`
- Create temporary vault directory structure
- Generate test identity using `age-keygen`

**Test Suite 1: Environment Variable Configuration**
- Set up environment variables:
  - `AGE_VAULT_IDENTITY_FILE` pointing to test identity
  - `AGE_VAULT_KEY_FILE` pointing to test vault key location
  - `AGE_VAULT_SSH_KEYS_DIR` pointing to test SSH keys directory
- Test workflow 1a: Initialize vault with env vars
  - Run `age-vault vault-key from-identity`
  - Verify vault key file exists
  - Verify it fails on second run (already exists)
- Test workflow 1b: Encrypt/decrypt data with env vars
  - Create test file with sample content
  - Run `age-vault encrypt` on test file
  - Run `age-vault decrypt` on encrypted file
  - Verify decrypted content matches original
- Test workflow 1c: Multi-user vault sharing with env vars
  - Generate second identity for "new user"
  - Get pubkey using `age-vault identity pubkey`
  - Encrypt vault key for new user using `age-vault vault-key encrypt`
  - Simulate new user: decrypt and verify with their identity
- Test workflow 1d: SSH agent with env vars
  - Generate test SSH keypair
  - Encrypt SSH private key using `age-vault encrypt`
  - Start SSH agent in background using `age-vault ssh start-agent`
  - Start test SSH server using `./test/start_ssh_server.sh` with the SSH public key
  - Test SSH connection using the agent (SSH_AUTH_SOCK)
  - Verify connection succeeds
  - Add a second SSH key to the directory while agent is running
  - Verify agent picks up the new key on next signing request
  - Stop SSH agent and server

**Test Suite 2: Config File Configuration**
- Create a test subdirectory with an `age_vault.yml` config file
- Set relative paths in the config file:
  - `vault_key_file: ./vault/vault_key.age`
  - `identity_file: ./vault/identity.txt`
  - `ssh_keys_dir: ./vault/ssh_keys`
- Unset all environment variables
- Change to the test subdirectory
- Test workflow 2a: Initialize vault with config file
  - Run `age-vault vault-key from-identity` from the subdirectory
  - Verify vault key is created relative to config file location (not cwd)
  - Verify paths resolve correctly
- Test workflow 2b: Encrypt/decrypt data with config file
  - Create test file with sample content
  - Run `age-vault encrypt` and `age-vault decrypt`
  - Verify operations work with config file paths
- Test workflow 2c: Change to different directory and verify paths still work
  - Create subdirectory within test directory
  - Change to that subdirectory
  - Run encrypt/decrypt operations
  - Verify config file is found by traversing up and paths resolve relative to config file location
- Test workflow 2d: SSH agent with config file
  - Generate and encrypt test SSH keypair
  - Start SSH agent using config file settings
  - Verify agent loads keys from correct relative path

**Test Suite 3: SOPS Passthrough**
- Use environment variable configuration for this suite
- Test workflow 3a: SOPS encrypt with age-vault
  - Create test YAML file with secrets
  - Run `age-vault sops -e test.yaml > test.enc.yaml`
  - Verify file is encrypted
- Test workflow 3b: SOPS decrypt with age-vault
  - Run `age-vault sops -d test.enc.yaml`
  - Verify decrypted content matches original
- Test workflow 3c: SOPS edit workflow
  - Create encrypted file
  - Run `age-vault sops test.enc.yaml` to edit
  - Verify SOPS can edit the file
- Test workflow 3d: SOPS with config file
  - Use config file configuration
  - Repeat encrypt/decrypt test
  - Verify SOPS passthrough works with config file settings

- Clean up temporary files and directories
- Print test results summary with pass/fail for each workflow

### flake.nix

**Modified: devShells.default.buildInputs**
- Add `openssh` to the list of development dependencies (needed for sshd and ssh-keygen in tests)
- The package already includes `age` and `sops`, which are needed
