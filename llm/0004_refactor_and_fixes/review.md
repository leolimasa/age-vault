# Go Files Review

## cmd/age-vault/main.go
- [x] main

## cmd/age-vault/commands/decrypt.go
- [x] RunDecrypt

## cmd/age-vault/commands/encrypt.go
- [x] RunEncrypt

## cmd/age-vault/commands/identity_pubkey.go
- [x] RunIdentityPubkey

## cmd/age-vault/commands/identity_set.go
- [x] RunIdentitySet

## cmd/age-vault/commands/vault_key_from_identity.go
- [x] RunVaultKeyFromIdentity

## cmd/age-vault/commands/vault_key_pubkey.go
- [x] RunVaultKeyPubkey

## cmd/age-vault/commands/vault_key_set.go
- [x] RunVaultKeySet

## cmd/age-vault/commands/vault_key_encrypt.go
- [x] RunVaultKeyEncrypt

## config/config.go
- [x] NewConfig
- [x] findAndLoadYAMLConfig
- [x] getConfigValue
- [x] expandHomePath
- [x] resolveConfigPath
- [x] EnsureParentDir


## keymgmt/keymgmt.go
- [x] LoadIdentity
- [x] LoadRecipient
- [x] ExtractRecipient
- [x] ExtractRecipientString
- [x] CopyFile
- [x] GetOrCreateVaultKey
- [x] SaveVaultKeyForIdentity


## vault/vault.go
- [x] GenerateVaultKey
- [x] EncryptVaultKey
- [x] DecryptVaultKey
- [x] (VaultKey) Encrypt
- [x] (VaultKey) Decrypt
- [x] (VaultKey) GetIdentity


# Integrations (2nd phase review)
## cmd/age-vault/commands/sops.go
- [ ] RunSops

## cmd/age-vault/commands/ssh_list_keys.go
- [ ] RunSSHListKeys

## cmd/age-vault/commands/ssh_start_agent.go
- [ ] RunSSHStartAgent


## sshagent/agent.go
- [ ] NewVaultSSHAgent
- [ ] (VaultSSHAgent) loadKey
- [ ] (VaultSSHAgent) List
- [ ] (VaultSSHAgent) Sign
- [ ] (VaultSSHAgent) SignWithFlags
- [ ] keysEqual
- [ ] (VaultSSHAgent) reloadKeys
- [ ] (VaultSSHAgent) Signers
- [ ] (VaultSSHAgent) Start
- [ ] (VaultSSHAgent) Add
- [ ] (VaultSSHAgent) Remove
- [ ] (VaultSSHAgent) RemoveAll
- [ ] (VaultSSHAgent) Lock
- [ ] (VaultSSHAgent) Unlock
- [ ] (VaultSSHAgent) Extension
