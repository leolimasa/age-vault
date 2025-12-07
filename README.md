# age-vault

`age-vault` lets one user encrypt data and all vault users decrypt it by encrypting a central **vault key** with each user's `age` public key.

## Usage

* `age-vault encrypt [file]`: encrypts a file using the vault key. Will read from stdin if no file is provided. Will output to stdout unless `-o [output file]` is provided.
* `age-vault decrypt [file]`: decrypts a file using the vault key. Will read from stdin if no file is provided. Will output to stdout unless `-o [output file]` is provided.
* `age-vault sops ...`: a passthrough to `sops` that sets up the vault key as an age identity before running sops commands. Example: `age-vault sops -d secrets.enc.yaml`. Requires `sops` to be installed.
* `age-vault key add [public key file]`: encrypts the vault key with the provided public key and returns the encrypted version. If a vault key does not yet exist (`AGE_VAULT_KEY` doesn't exist), one is created and then encrypted using the provided key. Will output to stdout unless `-o [output file]` is provided.
* `age-vault ssh-agent [key dir]`: starts an ssh-agent that loads vault encrypted SSH keys from the provided directory (or `AGE_VAULT_SSH_KEYS_DIR` if not provided). The keys should be encrypted using the vault key. The agent will decrypt them on demand using the vault key.

## Environment variables

`AGE_VAULT_KEY`: the age identity used to encrypt/decrypt files. If not set, the key is read from the file `~/.age-vault/vault_key.agekey` after decrypting it using the user's private key.
`AGE_VAULT_SSH_KEYS_DIR`: the directory containing vault encrypted SSH keys to be loaded by `age-vault ssh-agent`.

## New user/machine workflow

Follow this workflow to add a new user/machine to the vault:

* Have the user create a public/private key pair on their machine using `age` (ideally, in an HSM)
* Have the user send you the public key
* Run `age-vault key add [public key file]` to get the encrypted vault key for that user
* Send the encrypted vault key to the user to be stored in `~/.age-vault/vault_key.agekey`

## Motivation

I have several personal machines and need to share secrets between them without the overhead of a full service like hashicorp vault. 

Each machine stores its private key in an HSM (TPM on linux, yubikey on desktop, and secure enclave on mac). The private key never leaves the HSM.

The solution was to create a master key which is decrypted by the private keys of each machine's HSM and then in turn used to decrypt secrets. The master key is never decrypted to disk. It only ever exists decrypted in memory.

This is a very simple system, and was not designed for large teams or enterprises. It is meant for personal use or small teams where trust is not an issue. 

A key revocation mechanism is not present. Once a user is removed the entire vault needs to be re-encrypted with a new vault key. The easiest way to revoke a machine/user is to delete its private key from the HSM.
