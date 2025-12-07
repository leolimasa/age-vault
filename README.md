# age-vault

`age-vault` lets one user encrypt data and all vault users decrypt it by encrypting a central **vault key** with each user's `age` public key.

It provides an internal ssh agent that allows using ssh keys stored in the vault.

## Usage

* `age-vault encrypt [file]`: encrypts a file using the vault key. Will read from stdin if no file is provided.
* `age-vault decrypt [file]`: decrypts a file using the vault key. Will read from stdin if no file is provided.
* `age-vault user add [public key file]`: allows a public key to decrypt vault contents. If a vault key does not yet exist, one is created and then encrypted by the user's public key.
* `age-vault user remove [username]`: removes the user from the vault (deletes the user's encrypted vault key).
* `age-vault user list`: lists all users in the vault.
* `age-vault ssh-agent start`: starts a ssh agent process. Will read **vault encrypted** keys from `AGE_VAULT_SSH_KEYS_DIR`.
* `age-vault ssh-agent new-key`: creates a new ssh key pair, encrypts with the vault key, and stores in `AGE_VAULT_SSH_KEYS_DIR`.

## Environment variables

`AGE_VAULT_USER`: the current user using the vault. Defaults to `[username]@[hostname]`.
`AGE_VAULT_DIR`: directory where encrypted vault keys are stored. If not set, the program will traverse
`AGE_VAULT_SSH_KEYS_DIR`: directory where encrypted ssh keys are stored.

## Sops integration

TODO

## Architecture

