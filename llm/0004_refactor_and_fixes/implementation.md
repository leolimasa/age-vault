# Implementation Plan

## Data Structures

No new data structures will be created. Existing structures remain unchanged.

## Functions to Create

### keymgmt/keymgmt.go

#### `VaultKeyFromIdentityFile(identityFilePath string, vaultKeyFilePath string) (*vault.VaultKey, error)` (NEW)
Loads a user identity from file, reads the encrypted vault key from disk, decrypts it using the identity, and returns the decrypted vault key wrapped in a `vault.VaultKey` object. This consolidates the repeated pattern of loading identity → reading vault key file → decrypting vault key that appears in multiple commands.

**Implementation:**
- Call `LoadIdentity(identityFilePath)` to load the user identity
- Read the encrypted vault key bytes from `vaultKeyFilePath` using `os.ReadFile`
- Call `vault.DecryptVaultKey(encryptedBytes, userIdentity)` to decrypt the vault key
- Return the resulting `*vault.VaultKey`
- Handle errors at each step with descriptive messages

#### `NewClientUI() *plugin.ClientUI` (NEW)
Creates and returns a configured `plugin.ClientUI` instance that properly prompts the user for input when needed. This eliminates duplication of ClientUI creation across multiple functions and ensures consistent user interaction behavior.

**Implementation:**
- Create a `plugin.ClientUI` struct with the following callbacks:
  - `DisplayMessage`: Print plugin messages to stderr
  - `RequestValue`: Prompt user for input using a terminal library (e.g., `term` package) to support both regular and secret input
  - `Confirm`: Prompt user for yes/no confirmation with appropriate defaults
  - `WaitTimer`: Display waiting message to stderr
- Use the `golang.org/x/term` package to read passwords securely when `secret=true`
- Return the configured ClientUI

## Functions to Modify

### keymgmt/keymgmt.go

#### `LoadIdentity(path string) (age.Identity, error)` (MODIFY)
Simplify the plugin identity loading logic. Instead of checking if parsing as native identity fails first, attempt to parse as a native identity, and if that fails, immediately try loading as a plugin identity without pre-checking the file format.

**Changes:**
- Keep the native identity parsing attempt (lines 23-33)
- Remove the manual parsing of the identity file to find "AGE-PLUGIN-" prefix (lines 35-56)
- Instead, directly read the entire file content and pass it to `plugin.NewIdentity` after native parsing fails
- Replace the manually created ClientUI with a call to `NewClientUI()`
- Clean up error messages

#### `ExtractRecipient(identity age.Identity) (age.Recipient, error)` (NO CHANGES)
This function already correctly extracts recipients from any identity type. Keep as-is.

#### `ExtractRecipientString(identityPath string) (string, error)` (DELETE)
Remove this function entirely. The logic of extracting recipient strings should be done by calling `LoadIdentity` + `ExtractRecipient` + `.String()` where needed.

#### `LoadRecipient(path string) (age.Recipient, error)` (DELETE)
Remove this function entirely. Recipients should be derived from identities using `ExtractRecipient`, not loaded directly from files. When a user provides a public key explicitly, it will be passed as a string argument and parsed inline.

#### `GetOrCreateVaultKey(cfg *config.Config) (age.Identity, error)` (MODIFY)
Replace the manual identity loading and vault key decryption with a call to `VaultKeyFromIdentityFile`.

**Changes:**
- Keep the vault key existence check (lines 249-252)
- Replace lines 258-277 with:
  ```go
  vaultKey, err := VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
  if err != nil {
      return nil, fmt.Errorf("failed to load vault key: %w", err)
  }
  return vaultKey.GetIdentity()
  ```

#### `SaveVaultKeyForIdentity(vaultKeyIdentity age.Identity, userIdentity age.Identity, savePath string) (error)` (MODIFY)
Ensure this function works correctly with plugin identities. The function already uses `ExtractRecipient` which supports plugin identities, but verify it handles all identity types properly.

**Changes:**
- Review current implementation (lines 283-307)
- The current implementation already calls `ExtractRecipient(userIdentity)` which handles both native and plugin identities
- No changes needed - the function already supports plugin identities

### config/config.go

#### `findAndLoadYAMLConfig() (yamlConfig, string, error)` (MODIFY)
After traversing parent directories fails to find `age_vault.yml`, attempt to load a default config file from `~/.config/.age-vault/age_vault.yml`.

**Changes:**
- Keep the current parent directory traversal logic (lines 89-124)
- Before returning the empty config at line 124, add:
  ```go
  // Try default location
  homeDir, err := os.UserHomeDir()
  if err == nil {
      defaultConfigPath := filepath.Join(homeDir, ".config", ".age-vault", "age_vault.yml")
      if _, err := os.Stat(defaultConfigPath); err == nil {
          // File exists at default location
          data, err := os.ReadFile(defaultConfigPath)
          if err != nil {
              return cfg, "", fmt.Errorf("error reading default config file %s: %w", defaultConfigPath, err)
          }
          if err := yaml.Unmarshal(data, &cfg); err != nil {
              return cfg, "", fmt.Errorf("error parsing default config file %s: %w", defaultConfigPath, err)
          }
          return cfg, filepath.Join(homeDir, ".config", ".age-vault"), nil
      }
  }
  ```

### vault/vault.go

#### `DecryptVaultKey(encryptedKey []byte, userIdentity age.Identity) (*VaultKey, error)` (NO CHANGES)
This function already accepts `age.Identity` interface which supports both native and plugin identities. The `age.Decrypt` function handles plugin identities correctly. No changes needed.

### cmd/age-vault/commands/encrypt.go

#### `RunEncrypt(inputPath, outputPath string, cfg *config.Config) (error)` (MODIFY)
Replace the manual identity loading and vault key decryption with a call to `keymgmt.VaultKeyFromIdentityFile`.

**Changes:**
- Replace lines 17-32 with:
  ```go
  vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
  if err != nil {
      return fmt.Errorf("failed to load vault key: %w", err)
  }
  ```

### cmd/age-vault/commands/decrypt.go

#### `RunDecrypt(inputPath, outputPath string, cfg *config.Config) (error)` (MODIFY)
Replace the manual identity loading and vault key decryption with a call to `keymgmt.VaultKeyFromIdentityFile`.

**Changes:**
- Replace lines 16-33 with:
  ```go
  vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
  if err != nil {
      return fmt.Errorf("failed to load vault key: %w", err)
  }
  ```

### cmd/age-vault/commands/vault_key_pubkey.go

#### `RunVaultKeyPubkey(outputPath string, cfg *config.Config) (error)` (MODIFY)
Replace the manual identity loading and vault key decryption with a call to `keymgmt.VaultKeyFromIdentityFile`. Also use `ExtractRecipient` instead of type assertion.

**Changes:**
- Replace lines 17-39 with:
  ```go
  vaultKey, err := keymgmt.VaultKeyFromIdentityFile(cfg.IdentityFile, cfg.VaultKeyFile)
  if err != nil {
      return fmt.Errorf("failed to load vault key: %w", err)
  }

  vaultKeyIdentity, err := vaultKey.GetIdentity()
  if err != nil {
      return fmt.Errorf("failed to get vault key identity: %w", err)
  }

  recipient, err := keymgmt.ExtractRecipient(vaultKeyIdentity)
  if err != nil {
      return fmt.Errorf("failed to extract recipient from vault key: %w", err)
  }

  pubKeyStr := recipient.String() + "\n"
  ```

### cmd/age-vault/commands/identity_pubkey.go

#### `RunIdentityPubkey(outputPath string, cfg *config.Config) (error)` (MODIFY)
Replace the call to `ExtractRecipientString` with `LoadIdentity` + `ExtractRecipient`.

**Changes:**
- Replace lines 15-20 with:
  ```go
  identity, err := keymgmt.LoadIdentity(cfg.IdentityFile)
  if err != nil {
      return fmt.Errorf("failed to load identity: %w", err)
  }

  recipient, err := keymgmt.ExtractRecipient(identity)
  if err != nil {
      return fmt.Errorf("failed to extract recipient from identity: %w", err)
  }

  pubKeyStr := recipient.String() + "\n"
  ```

### cmd/age-vault/commands/vault_key_encrypt.go

#### `RunVaultKeyEncrypt(...) (error)` (MODIFY - signature and implementation change)
Change the function signature to accept multiple optional parameters for different ways of specifying the recipient. Support `--pubkey`, `--pubkey-file`, and `--identity` flags.

**New signature:**
```go
RunVaultKeyEncrypt(pubkey string, pubkeyFile string, identityFile string, outputPath string, cfg *config.Config) error
```

**Implementation:**
- Validate that exactly one of `pubkey`, `pubkeyFile`, or `identityFile` is provided
- If `pubkey` is provided: parse it using `age.ParseRecipients(strings.NewReader(pubkey))`
- If `pubkeyFile` is provided: read the file and parse using `age.ParseRecipients`
- If `identityFile` is provided: load identity using `keymgmt.LoadIdentity`, then extract recipient using `keymgmt.ExtractRecipient`
- Check if vault key exists; if not, generate a new one and save it encrypted for ourselves (keep existing logic from lines 20-57)
- If vault key exists, load it using `VaultKeyFromIdentityFile` (replaces lines 59-79)
- Encrypt the vault key for the target recipient
- Write to output

### cmd/age-vault/commands/vault_key_from_identity.go

#### `RunVaultKeyFromIdentity(cfg *config.Config) (error)` (DELETE)
Remove this command entirely. Its functionality is now part of `vault-key encrypt --identity`.

### cmd/age-vault/main.go

#### `main()` (MODIFY)
Remove the `vault-key from-identity` subcommand and update the `vault-key encrypt` subcommand to accept the new flags.

**Changes:**
- Remove lines 112-122 (vault-key from-identity command)
- Modify the vault-key encrypt command (lines 86-98) to add new flags:
  ```go
  var vaultKeyEncryptOutput string
  var vaultKeyEncryptPubkey string
  var vaultKeyEncryptPubkeyFile string
  var vaultKeyEncryptIdentity string

  vaultKeyEncryptCmd := &cobra.Command{
      Use:   "encrypt",
      Short: "Encrypt vault key for a new recipient",
      Long:  "Encrypts the vault key for a recipient. Use --pubkey, --pubkey-file, or --identity to specify the recipient.",
      Args:  cobra.NoArgs,
      RunE: func(cmd *cobra.Command, args []string) error {
          return commands.RunVaultKeyEncrypt(vaultKeyEncryptPubkey, vaultKeyEncryptPubkeyFile, vaultKeyEncryptIdentity, vaultKeyEncryptOutput, cfg)
      },
  }
  vaultKeyEncryptCmd.Flags().StringVar(&vaultKeyEncryptPubkey, "pubkey", "", "Public key string")
  vaultKeyEncryptCmd.Flags().StringVar(&vaultKeyEncryptPubkeyFile, "pubkey-file", "", "Path to public key file")
  vaultKeyEncryptCmd.Flags().StringVar(&vaultKeyEncryptIdentity, "identity", "", "Path to identity file (public key will be extracted)")
  vaultKeyEncryptCmd.Flags().StringVarP(&vaultKeyEncryptOutput, "output", "o", "", "Output file (default: stdout)")
  vaultKeyEncryptCmd.MarkFlagsMutuallyExclusive("pubkey", "pubkey-file", "identity")
  vaultKeyEncryptCmd.MarkFlagsOneRequired("pubkey", "pubkey-file", "identity")
  ```

## Documentation Updates

### README.md (MODIFY)
- Add documentation for `vault-key pubkey` command
- Update documentation for `vault-key encrypt` command to show the new flags (--pubkey, --pubkey-file, --identity)
- Remove documentation for `vault-key from-identity` command
- Add documentation about the default config file location at `~/.config/.age-vault/age_vault.yml`
