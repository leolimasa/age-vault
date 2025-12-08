# age-vault - High Level Requirements

## Project Overview
A command-line tool that enables secure secret sharing across multiple personal machines or within small teams using a centralized vault key system built on top of the `age` encryption tool.

## Core Requirements

### 1. Vault Key Management
- The system shall use a single master "vault key" that is used to encrypt/decrypt all secrets
- The vault key itself shall be encrypted using individual user public keys
- Each user shall be able to decrypt the vault key using their own private key stored in a Hardware Security Module (HSM)
- The vault key shall only exist decrypted in memory, never written to disk in plaintext
- Users shall be able to create, encrypt, and distribute the vault key to new users/machines

### 2. Data Encryption and Decryption
- Users shall be able to encrypt files using the vault key
- Users shall be able to decrypt files using the vault key
- The system shall support reading from stdin when no file is provided
- The system shall support writing to stdout unless an output file is specified
- All encryption/decryption operations shall automatically handle vault key decryption using the user's identity

### 3. Integration with External Tools
- The system shall provide a passthrough command for `sops` (Secrets OPerationS) that automatically configures the vault key as an age identity
- Users shall be able to run any `sops` command through age-vault without manual configuration

### 4. SSH Key Management
- The system shall provide an SSH agent implementation that manages vault-encrypted SSH keys
- The SSH agent shall automatically decrypt SSH keys on demand using the vault key
- Users shall be able to specify a directory containing encrypted SSH keys
- The SSH agent shall load all keys from the specified directory
- Users shall be able to list available SSH keys in the configured directory

### 5. User/Machine Onboarding
- The system shall support adding new users/machines to the vault
- New users shall be able to generate their own public/private key pairs using standard `age` tools
- Existing vault administrators shall be able to encrypt the vault key for new users
- The system shall provide commands to set up user identity and vault key files
- Users shall be able to extract the public key from their identity file using `age-vault identity pubkey`

### 6. Configuration Management
- The system shall use environment variables for configuration
- Default configuration locations shall be provided in `~/.config/.age-vault/`
- Users shall be able to override default locations via environment variables:
  - Vault key file location (`AGE_VAULT_KEY_FILE`)
  - User identity (private key) location (`AGE_VAULT_IDENTITY_FILE`)
  - SSH keys directory location (`AGE_VAULT_SSH_KEYS_DIR`)
- The system shall support YAML configuration files (`age_vault.yml`) for per-project configurations
- Configuration files shall be auto-detected by traversing up the directory tree from the current working directory
- YAML configuration shall support the following fields:
  - `vault_key_file` - Path to vault key file
  - `identity_file` - Path to identity file
  - `ssh_keys_dir` - Path to SSH keys directory

### 7. Security Requirements
- The vault key shall never be written to disk in decrypted form
- All secret operations shall happen in memory only

## Non-Requirements

### 1. User Revocation
- The system shall NOT provide an automated key revocation mechanism
- User removal is handled by deleting the private key from the HSM

### 2. Scale
- The system is NOT designed for large teams or enterprise use
- The system is intended for personal use or small teams with established trust
- No support for complex permission systems or role-based access control

## Target Use Case
Personal secret management across multiple machines where:
- Each machine has access to an HSM for private key storage
- Secrets need to be shared across all machines
- Full service secret management solutions (like HashiCorp Vault) are too complex/heavyweight
- All users/machines are trusted (no need for granular permissions)
