package tests

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/jacobsonjn/goaddress/internal/bootstrap"
	"github.com/stretchr/testify/assert"
)

const defaultContainerName = "address-parser"

// checkContainerRunning verifies if the address parser container is running
func checkContainerRunning() bool {
	checkCmd := exec.Command("docker", "ps", "--filter", "name="+defaultContainerName, "--format", "{{.Status}}")
	output, err := checkCmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(output)), "up")
}

func TestAddressParserHelp(t *testing.T) {
	bs := bootstrap.NewBootstrap()
	rootCmd := bs.Init()

	// Test cases for help and basic validation
	tests := []struct {
		name     string
		args     []string
		verbose  bool
		contains string // partial match for output
	}{
		{
			name:     "install help",
			args:     []string{"address", "install", "--help"},
			verbose:  false,
			contains: "Pull and start the address-parser-go-rest Docker container",
		},
		{
			name:     "parse help",
			args:     []string{"address", "parse", "--help"},
			verbose:  false,
			contains: "Parse an address into its components",
		},
		{
			name:     "expand help",
			args:     []string{"address", "expand", "--help"},
			verbose:  false,
			contains: "Expand an address into possible variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			output := buf.String()

			assert.NoError(t, err, "Help command should execute without error")
			assert.Contains(t, output, tt.contains)
		})
	}
}

func TestAddressParserCommands(t *testing.T) {
	// Skip parse/expand tests if container is not running
	if !checkContainerRunning() {
		t.Skip("Skipping parse/expand tests because address-parser container is not running")
	}

	bs := bootstrap.NewBootstrap()
	rootCmd := bs.Init()

	// Test cases for actual command execution
	tests := []struct {
		name     string
		args     []string
		verbose  bool
		wantErr  bool
		contains string // partial match for output or error
	}{
		{
			name:     "parse with no address",
			args:     []string{"address", "parse"},
			verbose:  false,
			wantErr:  true,
			contains: "address is required",
		},
		{
			name:     "expand with no address",
			args:     []string{"address", "expand"},
			verbose:  false,
			wantErr:  true,
			contains: "address is required",
		},
		{
			name:     "parse with address",
			args:     []string{"address", "parse", "123 Main St"},
			verbose:  false,
			wantErr:  false,
			contains: "{", // JSON response should at least contain opening brace
		},
		{
			name:     "expand with address",
			args:     []string{"address", "expand", "123 Main St"},
			verbose:  false,
			wantErr:  false,
			contains: "{", // JSON response should at least contain opening brace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			output := buf.String()

			if tt.wantErr {
				assert.Error(t, err, "Command should return error")
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err, "Command should execute without error")
				if tt.contains != "" {
					assert.Contains(t, output, tt.contains)
				}
			}
		})
	}
}
