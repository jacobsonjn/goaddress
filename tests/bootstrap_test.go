package tests

import (
	"bytes"
	"testing"

	"github.com/jacobsonjn/goaddress/internal/bootstrap"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBootstrap_Init(t *testing.T) {
	bs := bootstrap.NewBootstrap()
	rootCmd := bs.Init()

	assert.NotNil(t, rootCmd, "Root command should not be nil")
	assert.Equal(t, "toolboxcli", rootCmd.Use, "Root command name should be toolboxcli")

	// Check if subcommands are added
	subcommands := rootCmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 1, "At least one subcommand should be added")

	// Test greet subcommand existence
	var greetCmd *cobra.Command
	for _, cmd := range subcommands {
		if cmd.Use == "greet [name]" {
			greetCmd = cmd
			break
		}
	}
	assert.NotNil(t, greetCmd, "Greet subcommand should exist")
}

func TestBootstrap_Execute(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "without verbose",
			args:     []string{},
			expected: "Welcome to CLI Example! Use --help for usage.\n",
		},
		{
			name:     "with verbose",
			args:     []string{"--verbose"},
			expected: "Running in verbose mode\nWelcome to CLI Example! Use --help for usage.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := bootstrap.NewBootstrap()
			rootCmd := bs.Init()

			// Redirect output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// Execute root command
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()
			assert.NoError(t, err, "Root command execution should succeed")
			assert.Equal(t, tt.expected, buf.String(), "Root command should print expected message")
		})
	}
}
