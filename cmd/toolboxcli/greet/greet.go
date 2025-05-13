package greet

import (
	"fmt"

	"github.com/jacobsonjn/goaddress/internal/config"
	"github.com/spf13/cobra"
)

// NewGreetCommand creates a new greet subcommand.
func NewGreetCommand(cfg *config.Config) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Greet a user by name",
		Long:  `Prints a personalized greeting to the specified user.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Prioritize positional argument over flag
			if len(args) > 0 {
				name = args[0]
			}
			// If no name provided (via arg or flag), use default
			if name == "" {
				name = "Guest"
			}
			if cfg.Verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Verbose: Greeting %s\n", name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Hello, %s!\n", name)
		},
	}

	// Add flags specific to this command
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name to greet")

	return cmd
}
