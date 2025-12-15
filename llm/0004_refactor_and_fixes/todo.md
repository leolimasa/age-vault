# TODO Checklist

## Phase 1: Core Refactoring - keymgmt package

* [ ] Create `NewClientUI() *plugin.ClientUI` function in keymgmt/keymgmt.go
  * [ ] Implement DisplayMessage callback
  * [ ] Implement RequestValue callback with terminal input support
  * [ ] Implement Confirm callback with user prompting
  * [ ] Implement WaitTimer callback
* [ ] Create `VaultKeyFromIdentityFile(identityFilePath, vaultKeyFilePath string) (*vault.VaultKey, error)` in keymgmt/keymgmt.go
* [ ] Modify `LoadIdentity` in keymgmt/keymgmt.go to simplify plugin identity loading
  * [ ] Replace manual file parsing with direct plugin.NewIdentity call
  * [ ] Use NewClientUI() instead of inline ClientUI creation
* [ ] Delete `ExtractRecipientString` function from keymgmt/keymgmt.go
* [ ] Delete `LoadRecipient` function from keymgmt/keymgmt.go
* [ ] Modify `GetOrCreateVaultKey` in keymgmt/keymgmt.go to use VaultKeyFromIdentityFile
* [ ] Run tests for keymgmt package
* [ ] Validate implementation

## Phase 2: Config and Vault Updates

* [ ] Modify `findAndLoadYAMLConfig` in config/config.go to check default location ~/.config/.age-vault/age_vault.yml
* [ ] Run tests for config package
* [ ] Validate vault.DecryptVaultKey already supports plugin identities (no code changes needed)
* [ ] Validate keymgmt.SaveVaultKeyForIdentity already supports plugin identities (no code changes needed)
* [ ] Validate implementation

## Phase 3: Update Commands - Basic Operations

* [ ] Modify `RunEncrypt` in cmd/age-vault/commands/encrypt.go to use VaultKeyFromIdentityFile
* [ ] Modify `RunDecrypt` in cmd/age-vault/commands/decrypt.go to use VaultKeyFromIdentityFile
* [ ] Modify `RunVaultKeyPubkey` in cmd/age-vault/commands/vault_key_pubkey.go
  * [ ] Use VaultKeyFromIdentityFile
  * [ ] Use ExtractRecipient instead of type assertion
* [ ] Modify `RunIdentityPubkey` in cmd/age-vault/commands/identity_pubkey.go to use LoadIdentity + ExtractRecipient
* [ ] Run manual tests for encrypt, decrypt, vault-key pubkey, and identity pubkey commands
* [ ] Validate implementation

## Phase 4: Update vault-key encrypt Command

* [ ] Modify `RunVaultKeyEncrypt` signature in cmd/age-vault/commands/vault_key_encrypt.go to accept pubkey, pubkeyFile, identityFile parameters
* [ ] Implement validation that exactly one recipient source is provided
* [ ] Implement --pubkey flag handling (parse recipient from string)
* [ ] Implement --pubkey-file flag handling (read and parse recipient from file)
* [ ] Implement --identity flag handling (load identity and extract recipient)
* [ ] Update vault key loading logic to use VaultKeyFromIdentityFile when vault key exists
* [ ] Run manual tests for vault-key encrypt with each flag option
* [ ] Validate implementation

## Phase 5: Update Main CLI and Remove vault-key from-identity

* [ ] Modify main.go to update vault-key encrypt command definition
  * [ ] Add --pubkey, --pubkey-file, --identity flags
  * [ ] Mark flags as mutually exclusive
  * [ ] Mark one flag as required
  * [ ] Change Args from ExactArgs(1) to NoArgs
* [ ] Remove vault-key from-identity command from main.go
* [ ] Delete cmd/age-vault/commands/vault_key_from_identity.go file
* [ ] Delete cmd/age-vault/commands/vault_key_from_identity_test.go file if it exists
* [ ] Run manual tests to verify vault-key encrypt works with all three flag options
* [ ] Run manual tests to verify vault-key from-identity is no longer available
* [ ] Update test/run_integration_tests.sh with new test cases to cover all changes from all phases
* [ ] Run run_integration_tests.sh. Ensure all tests pass.
* [ ] Validate implementation

## Phase 6: Documentation

* [ ] Update README.md to add vault-key pubkey command documentation
* [ ] Update README.md to update vault-key encrypt command documentation with new flags
* [ ] Update README.md to remove vault-key from-identity command documentation
* [ ] Update README.md to document default config file location at ~/.config/.age-vault/age_vault.yml
* [ ] Validate implementation

## Phase 7: Final Testing and Cleanup

* [ ] Run all unit tests: `go test ./...`
* [ ] Build the project: `go build ./cmd/age-vault`
* [ ] Perform end-to-end testing with native age identities
* [ ] Perform end-to-end testing with plugin identities (if available)
* [ ] Review all code changes for consistency and style
* [ ] Validate implementation
