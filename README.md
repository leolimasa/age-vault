# age-vault

Share secrets across machines. One private key per machine. Keys never leaves the HSM. Built on top of the `age` encryption tool.

## Example

```bash

# Create age identity (use any age plugin) and initialize vault
age-plugin-tpm --generate -o age_identity.txt
age-vault vault-key from-identity age_identity.txt

# Encrypt a file
age-vault encrypt test_file.txt -o test_file.txt.age

# Decrypt a file
age-vault decrypt test_file.txt.age

# Add another machine/user to the vault (public key created on the other machine)
age-vault vault-key encrypt new_machine_public_key.txt -o new_machine_vault_key.txt

# Import a vault key (created by another machine that has access to the vault)
age-vault vault-key set new_machine_vault_key.txt
```

## Usage

| Command                               | Usage                                                                                                                                                                                         |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `age-vault encrypt [file]`            | Encrypts a file using the vault key. Reads from stdin if there is no output flag. Outputs to stdout unless `-o [output file]` is provided.                                                    |
| `age-vault decrypt [file]`            | Decrypts a file using the vault key. Reads from stdin if there is no output flag. Outputs to stdout unless `-o [output file]` is provided.                                                    |
| `age-vault sops ...`                  | A passthrough to `sops` that sets up the vault key as an age identity before running sops commands. Example: `age-vault sops -d secrets.enc.yaml`. Requires `sops` to be installed.           |
| `age-vault ssh start-agent [key dir]` | Starts an ssh-agent that loads vault encrypted SSH keys from the provided directory (or `AGE_VAULT_SSH_KEYS_DIR` if not provided). The agent will decrypt them on demand using the vault key. |
| `age-vault ssh list-keys`             | Lists the keys present in `AGE_VAULT_SSH_KEYS_DIR`                                                                                                                                            |

### Key management

* `age-vault vault-key from-identity`: creates a new vault key encrypted with the identity present in `AGE_VAULT_IDENTITY_FILE` and saves it to `AGE_VAULT_KEY_FILE`. Fails if a vault key already exists.
* `age-vault vault-key encrypt [public key file]`: encrypts the vault key with the provided public key and returns the encrypted version. If a vault key does not yet exist (`AGE_VAULT_KEY_FILE` doesn't exist), one is created and then encrypted using the provided key. Will output to stdout unless `-o [output file]` is provided.
* `age-vault vault-key set [encrypted key file]`: copies the provided encrypted vault key file to `AGE_VAULT_KEY_FILE`.
* `age-vault identity set [identity file]`: copies the identity file to the `AGE_VAULT_IDENTITY_FILE` localtion.
* `age-vault identity pubkey`: outputs the public key corresponding to the identity in `AGE_VAULT_IDENTITY_FILE`. Will output to stdout unless `-o [output file]` is provided.

## Config

**Supported environment variables:**

* `AGE_VAULT_KEY_FILE`: the vault key encrypted by the pubkey present in `AGE_VAULT_IDENTITY_FILE`. If not set, defaults to `~/.config/.age-vault/vault_key.age`.
* `AGE_VAULT_IDENTITY_FILE`: the age identity used to **encrypt and decrypt** the vault key. If not set, defaults to `~/.config/.age-vault/identity.txt`.
* `AGE_VAULT_SSH_KEYS_DIR`: the directory containing vault encrypted SSH keys to be loaded by `age-vault ssh-agent`.

**age_vault.yml config file:**

You can also create a config file `age_vault.yml` to set the above variables. It will be automatically detected by traversing up the directory tree from the current working directory. This makes it easy to have per-project vault configurations.

Example config file:

```yaml
vault_key_file: path/to/vault_key.age
identity_file: path/to/identity.txt
ssh_keys_dir: path/to/ssh_keys/
```

## New vault workflow

* Create a new age identity using one of the `age` keygen commands (like `age-plugin-tpm`)
* Move the newly created identity into the age vault using: `age-vault identity set [identity file]`
* Initialize the vault key using `age-vault vault-key from-identity`

## New user/machine workflow

Follow this workflow to add a new user/machine to the vault:

**On the target machine:**

* Create a public/private key pair using one of the `age` keygen commands (like `age-plugin-tpm`)
* Move the newly created identity into the age vault using: `age-vault identity set [identity file]`
* Get the public key with `age-vault identity pubkey -o [pubkeyfile]`, which will be used to encrypt the vault key for this machine.

**On another machine already setup with the vault:**

* Run `age-vault vault-key encrypt -o [user@machine.age] [pubkeyfile]` to get the encrypted vault key for that user/machine
* Send the encrypted vault key back to the target machine 

**Back on the target machine:**

* Load the vault key with `age-vault vault-key set [key file]`

## Vault key backup

Use the `age` command to backup the vault key (e.g to offline storage) by re-encrypting it using a passphrase:

```bash
age --decrypt -i "$AGE_VAULT_IDENTITY_FILE" "$AGE_VAULT_KEY_FILE" | age --passphrase -o **reencrypted_file.age**
```

## Motivation

I have several personal machines and need to share secrets between them without the overhead of a full service like hashicorp vault. 

Each machine stores its private key in an HSM (TPM on linux, yubikey on desktop, and secure enclave on mac). The private key never leaves the HSM.

The solution was to create a master key which is decrypted by the private keys of each machine's HSM and then in turn used to decrypt secrets. The master key is never decrypted to disk. It only ever exists decrypted in memory.

This is a very simple system, and was not designed for large teams or enterprises. It is meant for personal use or small teams where trust is not an issue. 

A key revocation mechanism is not present. Once a user is removed the entire vault needs to be re-encrypted with a new vault key. The easiest way to revoke a machine/user is to delete its private key from the HSM.
