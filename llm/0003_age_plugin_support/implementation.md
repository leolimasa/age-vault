# Age Plugin Support Implementation

## Overview
Add support for age plugin identities (like age-plugin-tpm) for user/machine identities, while keeping the vault key itself as X25519.

## Data Structures

No new data structures will be created. Existing structures will remain the same.

## Functions to be Modified

### vault/vault.go

#### `EncryptVaultKey(vaultKey age.Identity, recipientPubKey age.Recipient) ([]byte, error)` - MODIFIED
**Current behavior:** Only accepts X25519 recipient public keys.

**New behavior:** Will accept any age.Recipient, including plugin.Recipient types. The type assertion check for X25519Identity on the vault key will remain (vault key stays X25519), but the recipient parameter will be used as-is without type restrictions.

**Implementation approach:**
- Remove or relax any recipient type assertions
- The filippo.io/age library's `age.Encrypt()` function already accepts any `age.Recipient` interface, so it naturally supports plugin recipients
- Keep the vault key validation as X25519Identity (vault key itself doesn't use plugins)

#### `DecryptVaultKey(encryptedKey []byte, userIdentity age.Identity) (*VaultKey, error)` - MODIFIED
**Current signature uses:** `plugin.Identity` parameter type

**Current behavior:** Already accepts plugin.Identity which implements age.Identity interface, but we need to verify it works correctly with plugin identities.

**New behavior:** Will properly decrypt vault keys using any age.Identity including plugin-based identities. The age.Decrypt() function handles plugin identities automatically through the age.Identity interface.

**Implementation approach:**
- The current implementation already uses the correct interface types
- Verify that the identity parameter type is `age.Identity` (not restricted to X25519Identity)
- The filippo.io/age library's `age.Decrypt()` function handles plugin identities through the age.Identity interface

### keymgmt/keymgmt.go

#### `LoadIdentity(path string) (age.Identity, error)` - MODIFIED
**Current behavior:** Uses `age.ParseIdentities()` which only handles native age identities (X25519), not plugin identities.

**New behavior:** Will detect and load both native age identities and plugin identities.

**Implementation approach:**
- Try parsing as native identity first using `age.ParseIdentities()`
- If that fails or returns empty, attempt to parse as a plugin identity using `plugin.ParseIdentity()`
- Plugin identities are detected by their format (they start with plugin-specific prefixes like `AGE-PLUGIN-TPM-`)
- Return the first valid identity found, whether native or plugin-based

#### `LoadRecipient(path string) (age.Recipient, error)` - MODIFIED
**Current behavior:** Uses `age.ParseRecipients()` which only handles native age recipients (X25519 public keys), not plugin recipients.

**New behavior:** Will detect and load both native age recipients and plugin recipients.

**Implementation approach:**
- Try parsing as native recipient first using `age.ParseRecipients()`
- If that fails or returns empty, attempt to parse as a plugin recipient using `plugin.ParseRecipient()`
- Plugin recipients are detected by their format (they start with plugin-specific prefixes)
- Return the first valid recipient found, whether native or plugin-based

### cmd/age-vault/commands/vault_key_from_identity.go

#### `RunVaultKeyFromIdentity(cfg *config.Config) error` - MODIFIED
**Current behavior:** Type-asserts the loaded identity to `*age.X25519Identity` to extract the recipient, which fails for plugin identities.

**New behavior:** Will work with both X25519 identities and plugin identities by using a type-safe approach to extract recipients.

**Implementation approach:**
- Load the identity using the updated `keymgmt.LoadIdentity()` (which now supports plugins)
- Use a type switch or interface assertion to handle both X25519Identity and plugin.Identity types:
  - For X25519Identity: use `.Recipient()` method
  - For plugin.Identity: use `.Recipient()` method (plugin.Identity also has this method)
- Actually, both implement a way to get recipients - verify the exact interface and use it appropriately
- Alternatively, create a helper function `extractRecipient(identity age.Identity)` that handles this logic

#### `RunVaultKeyEncrypt(pubKeyPath, outputPath string, cfg *config.Config) error` - MODIFIED
**Current behavior:** Type-asserts the loaded identity to `*age.X25519Identity` to extract the recipient (in two places: when creating initial vault key for self, and when loading existing vault key).

**New behavior:** Will work with both X25519 identities and plugin identities.

**Implementation approach:**
- Update both locations where recipient extraction happens (lines ~36-40 and similar pattern later)
- Use the same recipient extraction approach as in `RunVaultKeyFromIdentity`
- The rest of the logic remains the same since `vault.EncryptVaultKey()` will accept any recipient type

### cmd/age-vault/commands/identity_pubkey.go

#### `RunIdentityPubkey(outputPath string, cfg *config.Config) error` - MODIFIED
**Current behavior:** Type-asserts the loaded identity to `*age.X25519Identity` to extract the recipient for display.

**New behavior:** Will work with both X25519 identities and plugin identities.

**Implementation approach:**
- Load identity using updated `keymgmt.LoadIdentity()`
- Use the same recipient extraction approach as in other modified commands
- Extract and display the public key string representation

## New Helper Function

### keymgmt/keymgmt.go

#### `ExtractRecipient(identity age.Identity) (age.Recipient, error)` - CREATED
**Purpose:** Safely extract a recipient (public key) from any age.Identity, whether native X25519 or plugin-based.

**Implementation approach:**
- Use type switch to handle different identity types:
  - `*age.X25519Identity`: call `.Recipient()` method
  - `*plugin.Identity`: call `.Recipient()` method
  - Default: return error for unsupported identity type
- This centralizes the recipient extraction logic used by multiple commands

## Refactoring for Code Reuse

The requirements mention that `vault_key_from_identity` and `vault_key_encrypt` share similar functionality. After reviewing the code:

**Shared functionality:**
1. Both load the user's identity
2. Both extract a recipient from the identity
3. Both generate or load a vault key
4. Both encrypt the vault key for a recipient
5. Both save encrypted vault key files

**Refactoring approach:**

### vault/vault.go or keymgmt/keymgmt.go

#### `GetOrCreateVaultKey(cfg *config.Config) (age.Identity, error)` - CREATED
**Purpose:** Load existing vault key if it exists, or generate a new one if it doesn't. Handles the "get or create" pattern used by both commands.

**Implementation approach:**
- Check if vault key file exists at `cfg.VaultKeyFile`
- If exists:
  - Load user identity via `keymgmt.LoadIdentity(cfg.IdentityFile)`
  - Read encrypted vault key file
  - Decrypt using `vault.DecryptVaultKey()`
  - Extract and return the vault key identity using `vaultKey.GetIdentity()`
- If doesn't exist:
  - Generate new vault key using `vault.GenerateVaultKey()`
  - Return the new vault key identity
- Note: This function does NOT save the vault key - that remains the caller's responsibility

#### `SaveVaultKeyForIdentity(vaultKeyIdentity age.Identity, userIdentity age.Identity, savePath string) error` - CREATED
**Purpose:** Encrypt a vault key for a specific identity and save it to disk.

**Implementation approach:**
- Extract recipient from userIdentity using `keymgmt.ExtractRecipient()`
- Encrypt vault key using `vault.EncryptVaultKey(vaultKeyIdentity, recipient)`
- Ensure parent directory exists using `config.EnsureParentDir(savePath)`
- Write encrypted bytes to file with 0600 permissions
- Return any errors encountered

These helper functions can then be used by both `RunVaultKeyFromIdentity` and `RunVaultKeyEncrypt` to reduce duplication.

## Testing Dependencies

### flake.nix - MODIFIED
**Current state:** Includes basic Go development tools and age utilities.

**Changes needed:**
- Add `age-plugin-tpm` to devShell buildInputs for testing
- Add `swtpm` package to devShell buildInputs (required for software TPM testing)
- Add `tpm2-tools` if needed for TPM operations

**Updated buildInputs section:**
```nix
buildInputs = with pkgs; [
  go
  gotools
  gopls
  go-tools
  golangci-lint
  age
  sops
  openssh
  age-plugin-tpm
  swtpm
  tpm2-tools
];
```
