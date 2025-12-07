# age-vault

`age-vault` lets one user encrypt data and all vault users decrypt it by encrypting a central **vault key** with each user's `age` public key.

It also provides an internal ssh agent that allows using ssh keys stored in the vault.


## Usage

* `age-vault encrypt [file]`: encrypts a file using the vault key. Will read from stdin if no file is provided.
* `age-vault decrypt [file]`: decrypts a file using the vault key. Will read from stdin if no file is provided.
* `age-vault user add [public key file]`: allows a public key to decrypt vault contents. If a vault key does not yet exist, one is created and then encrypted using the provided key.
* `age-vault ssh-agent start`: starts a ssh agent process. Will read **vault encrypted** keys from `AGE_VAULT_SSH_KEYS_DIR`.
* `age-vault ssh-agent new-key`: creates a new ssh key pair, encrypts with the vault key, and stores in `AGE_VAULT_SSH_KEYS_DIR`.

## Environment variables

`AGE_VAULT_USER`: the current user using the vault. Defaults to `[username]@[hostname]`.
`AGE_VAULT_DIR`: directory where encrypted vault keys are stored. If not set, the program will traverse all parent directories until it finds an `age_vault/` directory.
`AGE_VAULT_SSH_KEYS_DIR`: directory where encrypted ssh keys are stored.

## New user/machine workflow

Follow this workflow to add a new user to the vault:

* Have the user create a public/private key pair on their machine using `age` (ideally, in an HSM)
* Have the user send you the 

## Sops integration

TODO

## Motivation

I have several personal machines and need to share secrets between them without the overhead of a full service like hashicorp vault. 

Each machine stores its private key in an HSM (TPM on linux, yubikey on desktop, and secure enclave on mac). The private key never leaves the HSM.

I also need to be able to have a single SSH key that I can use from all those machines and that's also encrypted by the HSM. For my servers, I use a ssh ca (which gets rid of that problem), but some legacy services still require an SSH pub key.

The solution was to create a master key which is decrypted by the private keys of each machine's HSM and then in turn used to decrypt secrets. The master key is never decrypted to disk. It only ever exists decrypted in memory.

## Architecture

