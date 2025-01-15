package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sentinel-official/dvpn-node/cmd"
)

func main() {
	// Enable Cobra's feature to traverse and execute hooks for commands.
	cobra.EnableTraverseRunHooks = true

	// Retrieve the user's home directory.
	userDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to determine user home directory: %v\n", err)
		os.Exit(1)
	}

	// Define the default home directory for the application.
	homeDir := filepath.Join(userDir, ".dvpnx")

	// Initialize the root command for the application.
	rootCmd := cmd.NewRootCmd(homeDir)

	// Execute the root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
