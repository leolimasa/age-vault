# Age Plugin Support - Implementation Details

## Overview

The age-vault project now supports age plugin identities (like age-plugin-tpm, age-plugin-yubikey, etc.) for user/machine identities. The vault key itself remains X25519 for simplicity and compatibility.

## What Works

### ✅ Fully Functional

1. **Identity Loading**: Load both X25519 and plugin identities from files
2. **Recipient Loading**: Load both X25519 and plugin recipients (public keys) from files
3. **Public Key Extraction**: Extract public keys from plugin identities via comment parsing
4. **Backward Compatibility**: All existing X25519 functionality works unchanged

### ⚠️ Requires Plugin Binary Access

The following operations require the plugin binary (e.g., `age-plugin-tpm`) to have access to the underlying hardware (TPM, Yubikey, etc.):

1. **Vault Key Encryption**: Encrypting a vault key for a plugin recipient
2. **Vault Key Decryption**: Decrypting a vault key with a plugin identity
3. **File Operations**: Any encrypt/decrypt operations using plugin-based vault keys

## How It Works

### Identity File Format

Plugin identity files (e.g., from `age-plugin-tpm --generate`) contain:

```
# Created: 2025-12-12 ...
# Recipient: age1tpm1q...

AGE-PLUGIN-TPM-1Q...
```

Our implementation:
- Parses comment lines to extract the public key (recipient)
- Extracts the actual identity line (starting with `AGE-PLUGIN-`)
- Passes the identity to `plugin.NewIdentity()` for validation

### Plugin Execution Model

When you perform cryptographic operations with plugin identities:

1. **age-vault** calls the `filippo.io/age` library
2. The **age library** detects a plugin recipient/identity
3. The **age library** spawns the plugin binary (e.g., `age-plugin-tpm`)
4. The **plugin binary** communicates with the hardware (TPM, Yubikey, etc.)
5. Results are returned through the plugin protocol

**This means**: The plugin binary must be in your `$PATH` and have access to the hardware.

## Testing

### Automated Tests

Run the integration test suite:

```bash
./test_integration.sh
```

This tests:
- ✅ X25519 identity generation and usage (fully tested)
- ✅ Plugin identity parsing (tested)
- ✅ Plugin public key extraction (tested)
- ⚠️ Plugin vault operations (requires hardware - documented but not tested)

### Manual Testing with age-plugin-tpm

To fully test plugin support with a software TPM:

#### 1. Start swtpm (Software TPM)

```bash
mkdir -p /tmp/swtpm
swtpm socket --tpmstate dir=/tmp/swtpm --ctrl type=unixio,path=/tmp/swtpm/sock --tpm2 --daemon
```

#### 2. Set Environment Variable

```bash
export TPM2TOOLS_TCTI=swtpm:path=/tmp/swtpm/sock
```

#### 3. Generate Plugin Identity

```bash
age-plugin-tpm --generate --swtpm -o tpm_identity.txt
```

#### 4. Test Public Key Extraction (No TPM Required)

```bash
export AGE_VAULT_IDENTITY_FILE=tpm_identity.txt
age-vault identity pubkey
# Should output: age1tpm1q...
```

#### 5. Create Vault Key (Requires Running TPM)

```bash
export AGE_VAULT_IDENTITY_FILE=tpm_identity.txt
export AGE_VAULT_KEY_FILE=vault_key.age
age-vault vault-key from-identity
```

#### 6. Test Encryption/Decryption

```bash
echo "secret data" > test.txt
age-vault encrypt test.txt -o test.txt.age
age-vault decrypt test.txt.age
```

## Debugging

If operations hang or fail:

### Check Plugin Binary

```bash
which age-plugin-tpm
# Should show: /nix/store/.../bin/age-plugin-tpm
```

### Check TPM Access

```bash
# For swtpm:
export TPM2TOOLS_TCTI=swtpm:path=/tmp/swtpm/sock

# For hardware TPM:
export TPM2TOOLS_TCTI=device:/dev/tpmrm0

# Test access:
tpm2_getrandom 8 --hex
```

### Enable Debug Logging

The vault-key from-identity command now includes debug logging:

```bash
age-vault vault-key from-identity 2>&1 | grep DEBUG
```

Look for where it hangs:
- `[DEBUG] Creating encryptor...` - This is where the age library calls the plugin
- If it hangs here, the plugin is waiting for TPM access

## Architecture Decisions

### Why Keep Vault Key as X25519?

1. **Simplicity**: The vault key is shared among multiple users
2. **Performance**: X25519 is fast and well-tested
3. **Compatibility**: No dependency on specific hardware
4. **Security Model**: User identities can be plugin-based (hardware-backed), but the shared vault key is software-based

### Plugin Support Boundaries

- **User Identities**: Can be X25519 OR plugin-based
- **Vault Key**: Always X25519
- **Cryptographic Operations**: Delegated to the age library and plugins

## Known Limitations

1. **Hardware Dependency**: Plugin operations require hardware access
2. **Plugin Binary Required**: The plugin binary must be installed and in PATH
3. **Interactive Prompts Not Supported**: Our ClientUI implementation returns errors for interactive prompts
4. **Test Coverage**: Full end-to-end testing requires actual hardware or properly configured swtpm

## Code Structure

### Key Functions

- `keymgmt.LoadIdentity()`: Loads X25519 or plugin identities
- `keymgmt.LoadRecipient()`: Loads X25519 or plugin recipients
- `keymgmt.ExtractRecipient()`: Extracts recipient from any identity type
- `keymgmt.ExtractRecipientString()`: Gets public key string (reads from comments for plugins)
- `vault.EncryptVaultKey()`: Encrypts vault key for any recipient type
- `vault.DecryptVaultKey()`: Decrypts vault key with any identity type

### Integration Points

The code integrates with the `filippo.io/age/plugin` package:

- `plugin.NewIdentity()`: Parse plugin identity from string
- `plugin.NewRecipient()`: Parse plugin recipient from string
- `plugin.Identity.Recipient()`: Get recipient from plugin identity
- `plugin.ClientUI`: Callback interface for plugin communication

## Future Improvements

Potential enhancements (not currently implemented):

1. **Interactive Prompt Support**: Implement full ClientUI for interactive operations
2. **Plugin Binary Detection**: Better error messages when plugin binary is missing
3. **TPM Health Check**: Verify TPM access before attempting operations
4. **Plugin Testing Framework**: Mock plugin protocol for testing without hardware
