package tests

import (
	"bytes"
	"testing"

	"github.com/jacobsonjn/goaddress/cmd/toolboxcli/greet"
	"github.com/jacobsonjn/goaddress/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGreetCommand(t *testing.T) {
	cfg := config.NewConfig()

	// Test cases
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "with positional argument",
			args:     []string{"Alice"},
			expected: "Hello, Alice!\n",
		},
		{
			name:     "with name flag",
			args:     []string{"--name", "Bob"},
			expected: "Hello, Bob!\n",
		},
		{
			name:     "with no args or flags",
			args:     []string{},
			expected: "Hello, Guest!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for each test to avoid state interference
			cmd := greet.NewGreetCommand(cfg)

			// Redirect output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set arguments
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()
			assert.NoError(t, err, "Greet command should execute without error")
			assert.Equal(t, tt.expected, buf.String(), "Output should match expected greeting")
		})
	}
}
