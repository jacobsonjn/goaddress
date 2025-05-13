/*
The root command defines the base CLI command and its flags.
*/

package root

import (
	"fmt"

	"github.com/jacobsonjn/goaddress/internal/config"
	"github.com/spf13/cobra"
)

// NewRootCommand creates a new root command.
func NewRootCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toolboxcli",
		Short: "A simple extensible CLI built with Cobra",
		Long:  `A CLI application that demonstrates Cobra usage with a modular structure.`,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.Verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Running in verbose mode\n")
			}
			fmt.Fprintf(cmd.OutOrStdout(), `
                  _  _                     _  _ 
   _             | || |                   | |(_)
 _| |_ ___   ___ | || |__   ___ _   _ ____| | _ 
(_   _) _ \ / _ \| ||  _ \ / _ ( \ / ) ___) || |
  | || |_| | |_| | || |_) ) |_| ) X ( (___| || |
   \__)___/ \___/ \_)____/ \___(_/ \_)____)\_)_|
                                                
		   toolboxcli by jacobsonjn

Use --help for usage.
`)
		},
	}

	// Add persistent flags
	cmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, "Enable verbose output")

	return cmd
}
