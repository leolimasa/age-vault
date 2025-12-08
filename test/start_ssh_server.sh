#!/usr/bin/env bash
set -euo pipefail

# This script starts a temporary SSH daemon for testing purposes.
# It accepts a public key file as the first argument and allows authentication with that key.

if [ $# -lt 1 ]; then
    echo "Usage: $0 <public-key-file> [port]"
    echo "  public-key-file: Path to the SSH public key to authorize"
    echo "  port: Optional port number (default: 2222)"
    exit 1
fi

PUBLIC_KEY_FILE="$1"
PORT="${2:-2222}"

if [ ! -f "$PUBLIC_KEY_FILE" ]; then
    echo "Error: Public key file not found: $PUBLIC_KEY_FILE"
    exit 1
fi

# Create temporary directory for SSH server configuration
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Setting up temporary SSH server in $TEMP_DIR"

# Generate temporary SSH host keys
ssh-keygen -t rsa -f "$TEMP_DIR/ssh_host_rsa_key" -N '' -q
ssh-keygen -t ed25519 -f "$TEMP_DIR/ssh_host_ed25519_key" -N '' -q

# Create authorized_keys file with the provided public key
mkdir -p "$TEMP_DIR/authorized_keys_dir"
cp "$PUBLIC_KEY_FILE" "$TEMP_DIR/authorized_keys_dir/authorized_keys"
chmod 600 "$TEMP_DIR/authorized_keys_dir/authorized_keys"

# Create sshd_config
cat > "$TEMP_DIR/sshd_config" <<EOF
# Temporary SSH server configuration for testing
Port $PORT
ListenAddress 127.0.0.1
HostKey $TEMP_DIR/ssh_host_rsa_key
HostKey $TEMP_DIR/ssh_host_ed25519_key

# Authentication
PubkeyAuthentication yes
AuthorizedKeysFile $TEMP_DIR/authorized_keys_dir/authorized_keys
PasswordAuthentication no
ChallengeResponseAuthentication no
UsePAM no

# Logging
LogLevel DEBUG
SyslogFacility AUTH

# Security
PermitRootLogin no
StrictModes no
PidFile $TEMP_DIR/sshd.pid

# Other
UseDNS no
EOF

echo "Starting SSH server on 127.0.0.1:$PORT"
echo "You can connect with: ssh -p $PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $(whoami)@127.0.0.1"

# Start sshd in debug mode (foreground)
/usr/sbin/sshd -D -f "$TEMP_DIR/sshd_config" -e
