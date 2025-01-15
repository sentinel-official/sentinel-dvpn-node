package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sentinel-official/dvpn-node/config"
	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/node"
)

// StartCmd creates and returns a new Cobra command to start the node application.
func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Sentinel dVPN node",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the home directory from viper configuration.
			homeDir := viper.GetString("home")

			// Unmarshal the configuration.
			cfg := &config.Config{}
			if err := viper.Unmarshal(cfg); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			// Validate the configuration.
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			// Initialize the logger.
			logger, err := log.NewLogger(cmd.OutOrStderr(), cfg.Log.GetFormat(), cfg.Log.GetLevel())
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			log.SetLogger(logger)

			// Create and configure the application context.
			ctx := context.New().
				WithHomeDir(homeDir).
				WithInput(cmd.InOrStdin())

			// Setup the application context.
			if err := ctx.Setup(cfg); err != nil {
				return fmt.Errorf("failed to setup context: %w", err)
			}

			// Seal the context to prevent further modifications.
			ctx.Seal()

			// Create and setup the node.
			n := node.New()
			if err := n.Setup(ctx, cfg); err != nil {
				return fmt.Errorf("failed to setup node: %w", err)
			}

			// Channel to capture errors from the node.
			errChan := make(chan error, 1)

			// Start the node.
			if err := n.Start(ctx, errChan); err != nil {
				return fmt.Errorf("failed to start node: %w", err)
			}

			// Set up channel to capture OS signals.
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

			// Wait for a signal or an error.
			select {
			case <-signalChan:
			case err := <-errChan:
				log.Error(err.Error())
			}

			// Stop the node gracefully.
			if err := n.Stop(ctx); err != nil {
				return fmt.Errorf("failed to stop node: %w", err)
			}

			log.Info("Node stopped successfully")
			return nil
		},
	}

	return cmd
}
