# Age Plugin Support - Implementation TODO

## Phase 1: Core Library Support

* [ ] Update `keymgmt/keymgmt.go` - Add `ExtractRecipient(identity age.Identity)` helper function
  - Use type switch to handle both `*age.X25519Identity` and `*plugin.Identity`
  - Return the recipient from the appropriate `.Recipient()` method
  - Return error for unsupported identity types

* [ ] Update `keymgmt/keymgmt.go` - Modify `LoadIdentity()` to support plugin identities
  - First attempt to parse using `age.ParseIdentities()`
  - If that fails or returns empty, try `plugin.ParseIdentity()`
  - Return first valid identity found

* [ ] Update `keymgmt/keymgmt.go` - Modify `LoadRecipient()` to support plugin recipients
  - First attempt to parse using `age.ParseRecipients()`
  - If that fails or returns empty, try `plugin.ParseRecipient()`
  - Return first valid recipient found

* [ ] Update `vault/vault.go` - Verify `EncryptVaultKey()` signature
  - Ensure recipientPubKey parameter is `age.Recipient` (not restricted type)
  - Remove any unnecessary type assertions on the recipient
  - Keep vault key validation as X25519Identity

* [ ] Update `vault/vault.go` - Verify `DecryptVaultKey()` signature
  - Ensure userIdentity parameter is `age.Identity` (not restricted to specific type)
  - Verify it works with plugin identities through age.Identity interface

* [ ] Write unit tests for `ExtractRecipient()` helper function
  - Test with X25519Identity
  - Test with mock plugin.Identity (if possible)
  - Test error case with unsupported type

* [ ] Write unit tests for updated `LoadIdentity()`
  - Test loading native X25519 identity
  - Test loading plugin identity format

* [ ] Write unit tests for updated `LoadRecipient()`
  - Test loading native X25519 recipient
  - Test loading plugin recipient format

* [ ] validate implementation

## Phase 2: Command Updates

* [ ] Update `cmd/age-vault/commands/vault_key_from_identity.go`
  - Replace X25519Identity type assertion with call to `keymgmt.ExtractRecipient()`
  - Verify error handling is appropriate

* [ ] Update `cmd/age-vault/commands/vault_key_encrypt.go`
  - Replace first X25519Identity type assertion (around line 36-40) with `keymgmt.ExtractRecipient()`
  - Replace second X25519Identity type assertion (in the "vault exists" branch) with `keymgmt.ExtractRecipient()`
  - Verify error handling is appropriate

* [ ] Update `cmd/age-vault/commands/identity_pubkey.go`
  - Replace X25519Identity type assertion with call to `keymgmt.ExtractRecipient()`
  - Verify error handling is appropriate

* [ ] Build the project to verify compilation
  - Run `go build ./...`
  - Fix any compilation errors

* [ ] validate implementation

## Phase 3: Refactoring for Code Reuse

* [ ] Create `GetOrCreateVaultKey(cfg *config.Config)` helper in vault or keymgmt package
  - Implement vault key loading logic (if exists)
  - Implement vault key generation logic (if doesn't exist)
  - Return vault key identity without saving

* [ ] Create `SaveVaultKeyForIdentity(vaultKeyIdentity, userIdentity age.Identity, savePath string)` helper
  - Extract recipient using `keymgmt.ExtractRecipient()`
  - Encrypt vault key for recipient
  - Ensure parent directory exists
  - Save to file with secure permissions

* [ ] Refactor `RunVaultKeyFromIdentity()` to use new helpers
  - Use `GetOrCreateVaultKey()` to get vault key (should generate new since file doesn't exist)
  - Use `SaveVaultKeyForIdentity()` to save encrypted vault key
  - Maintain existing error messages and behavior

* [ ] Refactor `RunVaultKeyEncrypt()` to use new helpers
  - Use `GetOrCreateVaultKey()` for the get-or-create logic
  - Use `SaveVaultKeyForIdentity()` when saving vault key for self (if newly created)
  - Keep existing recipient encryption logic for the target public key

* [ ] Run existing tests to verify refactoring didn't break functionality
  - Run `go test ./...`
  - Fix any failing tests

* [ ] validate implementation

## Phase 4: Testing Dependencies and Manual Testing

* [ ] Update `flake.nix` to add testing dependencies
  - Add `age-plugin-tpm` to devShell buildInputs
  - Add `swtpm` to devShell buildInputs
  - Add `tpm2-tools` to devShell buildInputs
  - Update vendorHash if needed after go.mod changes

* [ ] Test flake.nix changes
  - Run `nix flake check` to verify flake is valid
  - Run `nix develop` to enter development shell
  - Verify `age-plugin-tpm`, `swtpm`, and `tpm2-tools` are available in shell

* [ ] Manual test: Generate TPM plugin identity
  - Run `age-plugin-tpm --generate --swtpm -o test_identity.txt`
  - Verify identity file is created

* [ ] Manual test: Initialize vault with plugin identity
  - Run `age-vault identity set test_identity.txt`
  - Run `age-vault vault-key from-identity`
  - Verify vault key file is created

* [ ] Manual test: Get public key from plugin identity
  - Run `age-vault identity pubkey -o test_pubkey.txt`
  - Verify public key is generated and has plugin format

* [ ] Manual test: Encrypt vault key for plugin recipient
  - Run `age-vault vault-key encrypt test_pubkey.txt -o test_vault_key.age`
  - Verify encrypted vault key is created

* [ ] Manual test: Encrypt and decrypt data with plugin-based vault
  - Create test file: `echo "test secret" > test_file.txt`
  - Run `age-vault encrypt test_file.txt -o test_file.txt.age`
  - Run `age-vault decrypt test_file.txt.age`
  - Verify decrypted content matches original

* [ ] Manual test: Multi-user scenario with plugin
  - Generate second TPM identity on "another machine"
  - Extract public key from second identity
  - Encrypt vault key for second identity
  - Simulate setting vault key on second machine
  - Verify second machine can decrypt vault-encrypted files

* [ ] validate implementation

## Phase 5: Documentation

* [ ] Update README.md example to show plugin usage (if needed)
  - The example already shows `age-plugin-tpm --generate`
  - Verify it matches our implementation

* [ ] Add any necessary code comments for plugin support
  - Document the plugin identity detection logic
  - Document the recipient extraction logic

* [ ] validate implementation
