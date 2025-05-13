package tests

import (
	"bytes"
	"testing"

	"github.com/jacobsonjn/goaddress/internal/bootstrap"
	"github.com/stretchr/testify/assert"
)

func TestGreetCommand(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		args     []string
		verbose  bool
		expected string
	}{
		{
			name:     "with positional argument",
			args:     []string{"greet", "Alice"},
			verbose:  false,
			expected: "Hello, Alice!\n",
		},
		{
			name:     "with name flag",
			args:     []string{"greet", "--name", "Bob"},
			verbose:  false,
			expected: "Hello, Bob!\n",
		},
		{
			name:     "with no args or flags",
			args:     []string{"greet"},
			verbose:  false,
			expected: "Hello, Guest!\n",
		},
		{
			name:     "with verbose and positional argument",
			args:     []string{"--verbose", "greet", "Alice"},
			verbose:  true,
			expected: "Verbose: Greeting Alice\nHello, Alice!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new bootstrap and root command
			bs := bootstrap.NewBootstrap()
			rootCmd := bs.Init()

			// Redirect output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// Set arguments (including subcommand)
			rootCmd.SetArgs(tt.args)

			// Execute root command
			err := rootCmd.Execute()
			assert.NoError(t, err, "Greet command should execute without error")
			assert.Equal(t, tt.expected, buf.String(), "Output should match expected greeting")
		})
	}
}

func TestGreetCommand_InvalidArgs(t *testing.T) {
	bs := bootstrap.NewBootstrap()
	rootCmd := bs.Init()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	rootCmd.SetArgs([]string{"greet", "Alice", "Bob"}) // Too many args
	err := rootCmd.Execute()
	assert.Error(t, err, "Greet command should fail with too many arguments")
}