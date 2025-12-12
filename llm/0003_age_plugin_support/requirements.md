* Objective: to add support for "age" user/machine identities that are managed by "age plugins".
* No need to support plugin identitis for the vault key itself.
* Read https://words.filippo.io/age-plugins/ and https://pkg.go.dev/filippo.io/age/plugin to understand how age plugins works
* Also read README.md to understand the overall structure of the program.
* Change vault/vault.go EncryptVaultKey so that it accepts a "filippo.io/age/plugin" Recipient. This will allow it to use a public key from a plugin identity.
* Change vault/vault.go DecryptVaultKey to accept a plugin.Identity. This will allow it to delegate encryption to a plugin.
* Change keymgmt.go LoadIdentity to support loading identity from plugin.
* `vault_key_from_identity` and `vault_key_encrypt` share similar functionality. Refactor it so that common functionality is only implemented once.
* Change vault key "from identity" and "encrypt" to support plugin identities
* Use "https://github.com/Foxboron/age-plugin-tpm" with "--swtpm" for testing. Generate a software only tpm key then create and encrypt a vault key using it with "vault-key from-identity". Ensure to add any needed dependencies to flake.nix.
