# TODO

## Project Status

* 🟡 Phase 1: Implement vault-key from-identity command - IMPLEMENTED
* 🟡 Phase 2: Fix age_vault.yml relative path resolution - IMPLEMENTED
* 🟡 Phase 3: Add SSH agent logging - IMPLEMENTED
* 🟡 Phase 4: Implement SSH agent key reloading - IMPLEMENTED
* 🟡 Phase 5: Create test infrastructure - IMPLEMENTED
* 🟡 Phase 6: Implement integration tests - Environment Variables - IMPLEMENTED
* 🟡 Phase 7: Implement integration tests - Config File - IMPLEMENTED
* 🟡 Phase 8: Implement integration tests - SOPS Passthrough - IMPLEMENTED
* 🟡 Phase 9: Update development environment - IMPLEMENTED

---

## Phase 1: Implement vault-key from-identity command

* [ ] Create `cmd/age-vault/commands/vault_key_from_identity.go`
* [ ] Implement `RunVaultKeyFromIdentity()` function
* [ ] Update `cmd/age-vault/main.go` to add the new subcommand
* [ ] Perform unit tests for vault-key from-identity command
* [ ] validate implementation

## Phase 2: Fix age_vault.yml relative path resolution

* [ ] Modify `config/config.go` Config struct to add `configFileDir` field
* [ ] Update `findAndLoadYAMLConfig()` to return the config file directory
* [ ] Create `resolveConfigPath()` helper function
* [ ] Update `NewConfig()` to use `resolveConfigPath()` for all path fields from YAML
* [ ] Perform unit tests for config path resolution with test config files
* [ ] validate implementation

## Phase 3: Add SSH agent logging

* [ ] Modify `sshagent.NewVaultSSHAgent()` to add key loading logs
* [ ] Modify `sshagent.Sign()` to add signing request logs
* [ ] Modify `sshagent.SignWithFlags()` to add signing request logs
* [ ] Perform manual test by running ssh agent and observing logs
* [ ] validate implementation

## Phase 4: Implement SSH agent key reloading

* [ ] Create `sshagent.reloadKeys()` method
* [ ] Modify `sshagent.Start()` to call `reloadKeys()` before serving each connection
* [ ] Perform manual test by starting agent, adding a key, and verifying it's loaded
* [ ] validate implementation

## Phase 5: Create test infrastructure

* [ ] Create `test/start_ssh_server.sh` script
* [ ] Make `test/start_ssh_server.sh` executable
* [ ] Test the SSH server script manually
* [ ] Create `test/run_integration_tests.sh` script
* [ ] Make `test/run_integration_tests.sh` executable
* [ ] validate implementation

## Phase 6: Implement integration tests - Environment Variables

* [ ] Implement Test Suite 1 setup: environment variable configuration
* [ ] Implement test workflow 1a: vault initialization with env vars
* [ ] Implement test workflow 1b: encrypt/decrypt data with env vars
* [ ] Implement test workflow 1c: multi-user vault sharing with env vars
* [ ] Implement test workflow 1d: SSH agent functionality with env vars
* [ ] Run Test Suite 1 and fix any issues
* [ ] validate implementation

## Phase 7: Implement integration tests - Config File

* [ ] Implement Test Suite 2 setup: config file with relative paths
* [ ] Implement test workflow 2a: vault initialization with config file
* [ ] Implement test workflow 2b: encrypt/decrypt data with config file
* [ ] Implement test workflow 2c: verify paths work from different directories
* [ ] Implement test workflow 2d: SSH agent with config file
* [ ] Run Test Suite 2 and fix any issues
* [ ] validate implementation

## Phase 8: Implement integration tests - SOPS Passthrough

* [ ] Implement Test Suite 3 setup: SOPS environment
* [ ] Implement test workflow 3a: SOPS encrypt with age-vault
* [ ] Implement test workflow 3b: SOPS decrypt with age-vault
* [ ] Implement test workflow 3c: SOPS edit workflow
* [ ] Implement test workflow 3d: SOPS with config file
* [ ] Run Test Suite 3 and fix any issues
* [ ] validate implementation

## Phase 9: Update development environment

* [ ] Add `openssh` to `flake.nix` devShell buildInputs
* [ ] Test that nix develop shell includes all required tools
* [ ] validate implementation
