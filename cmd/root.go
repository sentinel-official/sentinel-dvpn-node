package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/sentinel-official/sentinel-go-sdk/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd returns the main/root command for the Sentinel dVPN node CLI.
// It sets up persistent flags, config file reading, and attaches subcommands.
func NewRootCmd(homeDir string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dvpnx",
		Short: "Run and manage the Sentinel dVPN node",
		Long: `The Sentinel dVPN node software lets users join the decentralized VPN network on the Sentinel Hub blockchain,
providing secure, private, and censorship-resistant internet access while earning cryptocurrency rewards.
It integrates with Cosmos-SDK, supports robust configuration, and offers tools for key management and
node initialization, ensuring privacy, performance, and ease of use.`,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Retrieve a global viper instance
			v := viper.GetViper()

			// Bind persistent flags to viper
			if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
				return fmt.Errorf("failed to bind persistent flags: %w", err)
			}

			// Bind normal (non-persistent) flags to viper as well
			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("failed to bind flags: %w", err)
			}

			// Read the user-specified (or default) home directory
			homeDir := v.GetString("home")

			// Construct the expected path to the config file
			cfgPath := filepath.Join(homeDir, "config.toml")

			// Check if the config file exists; if not, that's okay—skip loading
			if _, err := os.Stat(cfgPath); err != nil {
				if os.IsNotExist(err) {
					return nil
				}

				return fmt.Errorf("failed to stat config file: %w", err)
			}

			// If the config file exists, set it and read it
			v.SetConfigFile(cfgPath)
			if err := v.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			return nil
		},
	}

	// Add subcommands
	rootCmd.AddCommand(
		cmd.KeysCmd(),
		InitCmd(),
		StartCmd(),
		version.NewVersionCommand(),
	)

	// Persistent flags
	rootCmd.PersistentFlags().String("home", homeDir, "Home directory for application config and data")

	return rootCmd
}
