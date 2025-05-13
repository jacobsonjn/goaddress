package addressparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/jacobsonjn/goaddress/internal/config"
	"github.com/spf13/cobra"
)

const (
	defaultContainerName = "address-parser"
	defaultContainerPort = "8080"
	defaultAPIEndpoint   = "http://localhost:8080"
)

// NewAddressParserCommand creates a new address-parser command group
func NewAddressParserCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Manage and use address parser service",
		Long:  `Install, manage and use address parser Docker container for address parsing and expansion.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newInstallCommand(cfg),
		newRemoveCommand(cfg),
		newParseCommand(cfg),
		newExpandCommand(cfg),
	)

	return cmd
}

func newInstallCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and start address parser container",
		Long:  `Pull and start the address-parser-go-rest Docker container.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "Installing address parser container...")
			}

			// Check if Docker is installed
			if err := exec.Command("docker", "--version").Run(); err != nil {
				return fmt.Errorf("docker not found. Please install Docker first")
			}

			// Check if container exists
			checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", defaultContainerName), "--format", "{{.Status}}")
			output, err := checkCmd.Output()
			containerStatus := string(output)
			containerExists := err == nil && len(containerStatus) > 0

			if containerExists {
				if !strings.Contains(strings.ToLower(containerStatus), "up") {
					// Container exists but is not running, start it
					startCmd := exec.Command("docker", "start", defaultContainerName)
					startCmd.Stdout = cmd.OutOrStdout()
					startCmd.Stderr = cmd.ErrOrStderr()
					if err := startCmd.Run(); err != nil {
						return fmt.Errorf("failed to start existing container: %v", err)
					}
					fmt.Fprintln(cmd.OutOrStdout(), "Address parser container started successfully!")
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Address parser container is already running!")
				return nil
			}

			// Container doesn't exist, pull the image and create new container
			pullCmd := exec.Command("docker", "pull", "registry.thedevforge.com/address-parser-go-rest:latest")
			pullCmd.Stdout = cmd.OutOrStdout()
			pullCmd.Stderr = cmd.ErrOrStderr()
			if err := pullCmd.Run(); err != nil {
				return fmt.Errorf("failed to pull Docker image: %v", err)
			}

			// Create and start new container
			startCmd := exec.Command("docker", "run", "-d",
				"--name", defaultContainerName,
				"-p", fmt.Sprintf("%s:8080", defaultContainerPort),
				"registry.thedevforge.com/address-parser-go-rest:latest",
			)
			startCmd.Stdout = cmd.OutOrStdout()
			startCmd.Stderr = cmd.ErrOrStderr()
			if err := startCmd.Run(); err != nil {
				return fmt.Errorf("failed to start container: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Address parser container installed and started successfully!")
			return nil
		},
	}

	return cmd
}

func newRemoveCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove address parser container",
		Long:  `Stop and remove the address-parser-go-rest Docker container.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "Removing address parser container...")
			}

			// Check if container exists
			checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", defaultContainerName), "--format", "{{.Status}}")
			output, err := checkCmd.Output()
			containerExists := err == nil && len(string(output)) > 0

			if !containerExists {
				fmt.Fprintln(cmd.OutOrStdout(), "Address parser container does not exist.")
				return nil
			}

			// Stop container if it's running
			stopCmd := exec.Command("docker", "stop", defaultContainerName)
			stopCmd.Stdout = cmd.OutOrStdout()
			stopCmd.Stderr = cmd.ErrOrStderr()
			if err := stopCmd.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to stop container: %v\n", err)
			}

			// Remove container
			rmCmd := exec.Command("docker", "rm", defaultContainerName)
			rmCmd.Stdout = cmd.OutOrStdout()
			rmCmd.Stderr = cmd.ErrOrStderr()
			if err := rmCmd.Run(); err != nil {
				return fmt.Errorf("failed to remove container: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Address parser container removed successfully!")
			return nil
		},
	}

	return cmd
}

// checkContainerRunning verifies if the address parser container is running
func checkContainerRunning() bool {
	checkCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", defaultContainerName), "--format", "{{.Status}}")
	output, err := checkCmd.Output()
	return err == nil && len(string(output)) > 0
}

func newParseCommand(cfg *config.Config) *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "parse [address]",
		Short: "Parse an address",
		Long:  `Parse an address into its components using the address parser service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				address = args[0]
			}
			if address == "" {
				return fmt.Errorf("address is required")
			}

			// Check if container is running
			if !checkContainerRunning() {
				return fmt.Errorf("address parser container is not running\nPlease run 'toolboxcli address install' first")
			}

			if cfg.Verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Parsing address: %s\n", address)
			}

			// Prepare request body
			payload := map[string]interface{}{
				"query": address,
			}
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to prepare request: %v", err)
			}

			// Send request to the service
			resp, err := http.Post(defaultAPIEndpoint+"/parse", "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				return fmt.Errorf("failed to connect to address parser service: %v\nMake sure the container is running with 'toolboxcli address install'", err)
			}
			defer resp.Body.Close()

			// Parse response
			var result interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			// Pretty print the result
			prettyJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format response: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "Address to parse")
	return cmd
}

func newExpandCommand(cfg *config.Config) *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "expand [address]",
		Short: "Expand an address",
		Long:  `Expand an address into possible variations using the address parser service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				address = args[0]
			}
			if address == "" {
				return fmt.Errorf("address is required")
			}

			// Check if container is running
			if !checkContainerRunning() {
				return fmt.Errorf("address parser container is not running\nPlease run 'toolboxcli address install' first")
			}

			if cfg.Verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Expanding address: %s\n", address)
			}

			// Prepare request body
			payload := map[string]interface{}{
				"query": address,
			}
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to prepare request: %v", err)
			}

			// Send request to the service
			resp, err := http.Post(defaultAPIEndpoint+"/expand", "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				return fmt.Errorf("failed to connect to address parser service: %v\nMake sure the container is running with 'toolboxcli address install'", err)
			}
			defer resp.Body.Close()

			// Parse response
			var result interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			// Pretty print the result
			prettyJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format response: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "Address to expand")
	return cmd
}
