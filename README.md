# age-vault

A command-line tool that enables secure secret sharing across multiple machines using a centralized vault key system built on top of the `age` encryption tool.

## Usage

* `age-vault encrypt [file]`: encrypts a file using the vault key. Will read from stdin if no file is provided. Will output to stdout unless `-o [output file]` is provided.
* `age-vault decrypt [file]`: decrypts a file using the vault key. Will read from stdin if no file is provided. Will output to stdout unless `-o [output file]` is provided.
* `age-vault sops ...`: a passthrough to `sops` that sets up the vault key as an age identity before running sops commands. Example: `age-vault sops -d secrets.enc.yaml`. Requires `sops` to be installed.
* `age-vault ssh-agent [key dir]`: starts an ssh-agent that loads vault encrypted SSH keys from the provided directory (or `AGE_VAULT_SSH_KEYS_DIR` if not provided). The agent will decrypt them on demand using the vault key.

### Key management

* `age-vault encrypt-vault-key [public key file]`: encrypts the vault key with the provided public key and returns the encrypted version. If a vault key does not yet exist (`AGE_VAULT_KEY_FILE` doesn't exist), one is created and then encrypted using the provided key. Will output to stdout unless `-o [output file]` is provided.
* `age-vault set-vault-key [vault key file]`: copies the vault key file to `AGE_VAULT_KEY_FILE`.
* `age-vault set-idenity [identity file]`: copies the identity file to the `AGE_VAULT_IDENTITY_FILE` localtion.
* `age-vault set-pubkey [public key file]`: copies the public key file to the `AGE_VAULT_PUBKEY_FILE` localtion.

## Environment variables

`AGE_VAULT_KEY_FILE`: the vault key encrypted by `AGE_VAULT_PUBKEY_FILE`. If not set, defaults to `~/.config/.age-vault/vault_key.age`.
`AGE_VAULT_PUBKEY_FILE`: the age public key used to **encrypt** the vault key. If not set, defaults to `~/.config/.age-vault/pubkey.txt`.
`AGE_VAULT_IDENTITY_FILE`: the age identity used to **decrypt** the vault key. If not set, defaults to `~/.config/.age-vault/identity.txt`.
`AGE_VAULT_SSH_KEYS_DIR`: the directory containing vault encrypted SSH keys to be loaded by `age-vault ssh-agent`.

## New user/machine workflow

Follow this workflow to add a new user/machine to the vault:

* Have the user create a public/private key pair on their machine using one of the `age` keygen commands (like `age-plugin-tpm`)
* Have the user move the newly created identity and public key into the age vault locations using:
  * `age-vault set-idenity [identity file]`
  * `age-vault set-pubkey [public key file]`
* Have the user send you the public key
* Run `age-vault encrypt-vault-key [public key file]` to get the encrypted vault key for that user
* Send the encrypted vault key to the user and load it into their `AGE_VAULT_KEY_FILE` location with `age-vault set-vault-key [key file]`

## Vault key backup

If you need to backup the vault key (e.g. to offline storage), you can use the `age` command to re-encrypt it using a passphrase:

```bash
age --decrypt -i "$AGE_VAULT_IDENTITY_FILE" "$AGE_VAULT_KEY_FILE" | age --passphrase -o **reencrypted_file.age**
```

## Motivation

I have several personal machines and need to share secrets between them without the overhead of a full service like hashicorp vault. 

Each machine stores its private key in an HSM (TPM on linux, yubikey on desktop, and secure enclave on mac). The private key never leaves the HSM.

The solution was to create a master key which is decrypted by the private keys of each machine's HSM and then in turn used to decrypt secrets. The master key is never decrypted to disk. It only ever exists decrypted in memory.

This is a very simple system, and was not designed for large teams or enterprises. It is meant for personal use or small teams where trust is not an issue. 

A key revocation mechanism is not present. Once a user is removed the entire vault needs to be re-encrypted with a new vault key. The easiest way to revoke a machine/user is to delete its private key from the HSM.
