
* Create a new function `keymgmt.VaultKeyFromIdentityFile(..)` that loads the user identity from a file (keymgmt.LoadIdentity), reads the encrypted vault key, decrypts it using the identity, and returns the decrypted version as a vault.VaultKey.
* Change encrypt.go, decrypt.go, vault_key_pubkey.go to use `keymgmt.VaultKeyFromIdentityFile`
* Remove the "vault-key from-identity" command. Instead, make the "vault-key encrypt" take these flags: "--pubkey [string]", that encrypts the vault key with a given public key, "--pubkey-file [path]" that encrypts the vault key with a public key from a file, or "--identity [file]" that extracts the public key from an age identity and uses that to encrypt the vault key. Update README.md accordingly.
* Delete ExtractRecipientString from keymgmt. Public keys should always be derived from private, unless explicitely given by the user as an argument.
* commands/identity_pubkey extracts the public key string from the identity file. Use ExtractRecipient instead
* add `vault-key pubkey` to the README.md
* vault_key_pubkey.go should use `ExtractRecipient`
* on config.go, if it can't find a yml file by tracing the parents, try to open a default file from "~/.config/.age-vault/age_vault.yml". Update the README.md accordingly.
* on keymgmt.go, have a common function that creates a `plugin.ClientUI` and use that everywhere that a `ClientUI` is needed.
	* The `ClientUI` MUST prompt the user if needed.
* Delete `LoadRecipient` from keymgmt.go. All recipients should come from the private key by using `ExtractRecipient`
* On keymgmt.go `LoadIdentity`, don't try to parse a plugin identity file to check if it's a plugin identity. Just go ahead and load it as a plugin anyways, because there are no other alternatives other than it being a plugin.
* On keymgmt.go, `GetOrCreateVaultKey` should use `VaultKeyFromIdentityFile`.
* On keymgmt.go, `SaveVaultKeyForIdentity` should support a plugin userIdentity.
* On vault.go, `DecryptVaultKey` should support a plugin userIdentity
