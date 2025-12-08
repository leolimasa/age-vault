package main

import (
	"fmt"
	"os"

	"github.com/leolimasa/age-vault/cmd/age-vault/commands"
	"github.com/leolimasa/age-vault/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

func main() {
	// Load configuration
	var err error
	cfg, err = config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "age-vault",
		Short: "A secure secret sharing tool built on age encryption",
		Long: `age-vault enables secure secret sharing across multiple machines using a
centralized vault key system. Built on top of the age encryption tool.`,
	}

	// Add encrypt command
	var encryptOutputFile string
	encryptCmd := &cobra.Command{
		Use:   "encrypt [file]",
		Short: "Encrypt a file using the vault key",
		Long:  "Encrypts a file using the vault key. Reads from stdin if no file is provided.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := ""
			if len(args) > 0 {
				inputPath = args[0]
			}
			return commands.RunEncrypt(inputPath, encryptOutputFile, cfg)
		},
	}
	encryptCmd.Flags().StringVarP(&encryptOutputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(encryptCmd)

	// Add decrypt command
	var decryptOutputFile string
	decryptCmd := &cobra.Command{
		Use:   "decrypt [file]",
		Short: "Decrypt a file using the vault key",
		Long:  "Decrypts a file using the vault key. Reads from stdin if no file is provided.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := ""
			if len(args) > 0 {
				inputPath = args[0]
			}
			return commands.RunDecrypt(inputPath, decryptOutputFile, cfg)
		},
	}
	decryptCmd.Flags().StringVarP(&decryptOutputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(decryptCmd)

	// Add sops passthrough command
	sopsCmd := &cobra.Command{
		Use:                "sops [sops-args...]",
		Short:              "Run sops with vault key",
		Long:               "Passthrough to sops that sets up the vault key as an age identity.",
		DisableFlagParsing: true, // Let sops handle its own flags
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunSops(args, cfg)
		},
	}
	rootCmd.AddCommand(sopsCmd)

	// Add vault-key command group
	vaultKeyCmd := &cobra.Command{
		Use:   "vault-key",
		Short: "Manage the vault key",
		Long:  "Commands for managing the vault key (create, encrypt for users, set)",
	}
	rootCmd.AddCommand(vaultKeyCmd)

	// Add vault-key encrypt subcommand
	var vaultKeyEncryptOutput string
	vaultKeyEncryptCmd := &cobra.Command{
		Use:   "encrypt [public-key-file]",
		Short: "Encrypt vault key for a new user",
		Long:  "Encrypts the vault key for a recipient's public key. Creates a new vault key if none exists.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunVaultKeyEncrypt(args[0], vaultKeyEncryptOutput, cfg)
		},
	}
	vaultKeyEncryptCmd.Flags().StringVarP(&vaultKeyEncryptOutput, "output", "o", "", "Output file (default: stdout)")
	vaultKeyCmd.AddCommand(vaultKeyEncryptCmd)

	// Add vault-key set subcommand
	vaultKeySetCmd := &cobra.Command{
		Use:   "set [encrypted-key-file]",
		Short: "Set the vault key from an encrypted file",
		Long:  "Copies the encrypted vault key file to the configured vault key location.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunVaultKeySet(args[0], cfg)
		},
	}
	vaultKeyCmd.AddCommand(vaultKeySetCmd)

	// Add identity command group
	identityCmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage user identity",
		Long:  "Commands for managing the user's identity (private key)",
	}
	rootCmd.AddCommand(identityCmd)

	// Add identity set subcommand
	identitySetCmd := &cobra.Command{
		Use:   "set [identity-file]",
		Short: "Set the identity from a file",
		Long:  "Copies the identity file to the configured identity location.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunIdentitySet(args[0], cfg)
		},
	}
	identityCmd.AddCommand(identitySetCmd)

	// Add identity pubkey subcommand
	var identityPubkeyOutput string
	identityPubkeyCmd := &cobra.Command{
		Use:   "pubkey",
		Short: "Output the public key for the identity",
		Long:  "Extracts and outputs the public key from the configured identity file.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunIdentityPubkey(identityPubkeyOutput, cfg)
		},
	}
	identityPubkeyCmd.Flags().StringVarP(&identityPubkeyOutput, "output", "o", "", "Output file (default: stdout)")
	identityCmd.AddCommand(identityPubkeyCmd)

	// Add ssh command group
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage SSH keys and agent",
		Long:  "Commands for managing vault-encrypted SSH keys and running the SSH agent",
	}
	rootCmd.AddCommand(sshCmd)

	// Add ssh start-agent subcommand
	var sshKeysDir string
	sshStartAgentCmd := &cobra.Command{
		Use:   "start-agent [keys-dir]",
		Short: "Start SSH agent with vault-encrypted keys",
		Long:  "Starts an SSH agent that loads vault-encrypted SSH keys from the specified directory.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keysDir := sshKeysDir
			if len(args) > 0 {
				keysDir = args[0]
			}
			return commands.RunSSHStartAgent(keysDir, cfg)
		},
	}
	sshStartAgentCmd.Flags().StringVar(&sshKeysDir, "keys-dir", "", "Directory containing encrypted SSH keys")
	sshCmd.AddCommand(sshStartAgentCmd)

	// Add ssh list-keys subcommand
	sshListKeysCmd := &cobra.Command{
		Use:   "list-keys",
		Short: "List encrypted SSH keys",
		Long:  "Lists all encrypted SSH keys in the configured keys directory.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunSSHListKeys(cfg)
		},
	}
	sshCmd.AddCommand(sshListKeysCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
