/*
A simple config struct to manage CLI settings (e.g., environment variables).
This makes the CLI extensible for future configuration needs.
*/

package config

import (
	"github.com/spf13/viper"
)

// Config holds CLI configuration.
type Config struct {
	Verbose bool
}

// NewConfig initializes configuration with defaults or environment variables.
func NewConfig() *Config {
	v := viper.New()
	v.SetDefault("verbose", false)
	v.SetEnvPrefix("CLI")
	v.AutomaticEnv()

	return &Config{
		Verbose: v.GetBool("verbose"),
	}
}
