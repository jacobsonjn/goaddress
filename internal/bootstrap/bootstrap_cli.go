/*
The Bootstrap struct initializes the Cobra root command and manages subcommands.
It’s designed to be extensible and testable by allowing dependency injection (e.g., config or logger).
*/
package bootstrap

import (
	"github.com/jacobsonjn/goaddress/cmd/toolboxcli/greet"
	"github.com/jacobsonjn/goaddress/cmd/toolboxcli/root"
	"github.com/jacobsonjn/goaddress/internal/config"
	"github.com/spf13/cobra"
)

// Bootstrap holds the CLI configuration and root command.
type BootstrapCli struct {
	rootCmd *cobra.Command
	config  *config.Config
}

// NewBootstrap creates a new Bootstrap instance with default configuration.
func NewBootstrap() *BootstrapCli {
	cfg := config.NewConfig()
	return &BootstrapCli{
		config: cfg,
	}
}

// Init initializes the Cobra root command and adds subcommands.
func (b *BootstrapCli) Init() *cobra.Command {

	// Initialize root command
	b.rootCmd = root.NewRootCommand(b.config)

	// ADD NEW COMMANDS HERE:
	b.rootCmd.AddCommand(greet.NewGreetCommand(b.config))

	return b.rootCmd
}

// For testing purposes, allow access to the root command
func (b *BootstrapCli) GetRootCmd() *cobra.Command {
	return b.rootCmd
}
