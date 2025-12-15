# TODO Checklist

## Project Status

* 🟢 Phase 1: Core Refactoring - keymgmt package - COMMITTED
* 🟢 Phase 2: Config and Vault Updates - COMMITTED
* 🟢 Phase 3: Update Commands - Basic Operations - COMMITTED
* 🟢 Phase 4: Update vault-key encrypt Command - COMMITTED
* 🟢 Phase 5: Update Main CLI and Remove vault-key from-identity - COMMITTED
* 🟢 Phase 6: Documentation - COMMITTED
* 🟢 Phase 7: Final Testing and Cleanup - COMMITTED

## Phase 1: Core Refactoring - keymgmt package

* [x] Create `NewClientUI() *plugin.ClientUI` function in keymgmt/keymgmt.go
  * [x] Implement DisplayMessage callback
  * [x] Implement RequestValue callback with terminal input support
  * [x] Implement Confirm callback with user prompting
  * [x] Implement WaitTimer callback
* [x] Create `VaultKeyFromIdentityFile(identityFilePath, vaultKeyFilePath string) (*vault.VaultKey, error)` in keymgmt/keymgmt.go
* [x] Modify `LoadIdentity` in keymgmt/keymgmt.go to simplify plugin identity loading
  * [x] Replace manual file parsing with direct plugin.NewIdentity call
  * [x] Use NewClientUI() instead of inline ClientUI creation
* [x] Delete `ExtractRecipientString` function from keymgmt/keymgmt.go
* [x] Delete `LoadRecipient` function from keymgmt/keymgmt.go
* [x] Modify `GetOrCreateVaultKey` in keymgmt/keymgmt.go to use VaultKeyFromIdentityFile
* [x] Run tests for keymgmt package
* [x] Validate implementation

## Phase 2: Config and Vault Updates

* [x] Modify `findAndLoadYAMLConfig` in config/config.go to check default location ~/.config/.age-vault/age_vault.yml
* [x] Run tests for config package
* [x] Validate vault.DecryptVaultKey already supports plugin identities (no code changes needed)
* [x] Validate keymgmt.SaveVaultKeyForIdentity already supports plugin identities (no code changes needed)
* [x] Validate implementation

## Phase 3: Update Commands - Basic Operations

* [x] Modify `RunEncrypt` in cmd/age-vault/commands/encrypt.go to use VaultKeyFromIdentityFile
* [x] Modify `RunDecrypt` in cmd/age-vault/commands/decrypt.go to use VaultKeyFromIdentityFile
* [x] Modify `RunVaultKeyPubkey` in cmd/age-vault/commands/vault_key_pubkey.go
  * [x] Use VaultKeyFromIdentityFile
  * [x] Use ExtractRecipient instead of type assertion
* [x] Modify `RunIdentityPubkey` in cmd/age-vault/commands/identity_pubkey.go to use LoadIdentity + ExtractRecipient
* [x] Run manual tests for encrypt, decrypt, vault-key pubkey, and identity pubkey commands
* [x] Validate implementation

## Phase 4: Update vault-key encrypt Command

* [x] Modify `RunVaultKeyEncrypt` signature in cmd/age-vault/commands/vault_key_encrypt.go to accept pubkey, pubkeyFile, identityFile parameters
* [x] Implement validation that exactly one recipient source is provided
* [x] Implement --pubkey flag handling (parse recipient from string)
* [x] Implement --pubkey-file flag handling (read and parse recipient from file)
* [x] Implement --identity flag handling (load identity and extract recipient)
* [x] Update vault key loading logic to use VaultKeyFromIdentityFile when vault key exists
* [x] Run manual tests for vault-key encrypt with each flag option
* [x] Validate implementation

## Phase 5: Update Main CLI and Remove vault-key from-identity

* [x] Modify main.go to update vault-key encrypt command definition
  * [x] Add --pubkey, --pubkey-file, --identity flags
  * [x] Mark flags as mutually exclusive
  * [x] Mark one flag as required
  * [x] Change Args from ExactArgs(1) to NoArgs
* [x] Remove vault-key from-identity command from main.go
* [x] Delete cmd/age-vault/commands/vault_key_from_identity.go file
* [x] Delete cmd/age-vault/commands/vault_key_from_identity_test.go file if it exists
* [x] Run manual tests to verify vault-key encrypt works with all three flag options
* [x] Run manual tests to verify vault-key from-identity is no longer available
* [x] Update test/run_integration_tests.sh with new test cases to cover all changes from all phases
* [x] Run run_integration_tests.sh. Ensure all tests pass.
* [x] Validate implementation

## Phase 6: Documentation

* [x] Update README.md to add vault-key pubkey command documentation
* [x] Update README.md to update vault-key encrypt command documentation with new flags
* [x] Update README.md to remove vault-key from-identity command documentation
* [x] Update README.md to document default config file location at ~/.config/.age-vault/age_vault.yml
* [x] Validate implementation

## Phase 7: Final Testing and Cleanup

* [x] Run all unit tests: `go test ./...`
* [x] Build the project: `go build ./cmd/age-vault`
* [x] Perform end-to-end testing with native age identities
* [x] Perform end-to-end testing with plugin identities (if available)
* [x] Review all code changes for consistency and style
* [x] Validate implementation
