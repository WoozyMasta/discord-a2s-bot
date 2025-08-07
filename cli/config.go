package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"
)

/*
Config represents the configuration settings for the application.

It includes bot settings, a list of servers to monitor,
logging configuration, and other relevant parameters.
*/
type Config struct {
	Logging Logging        `yaml:"logging,omitempty"` // Logging configuration
	Servers []ServerConfig `yaml:"servers"`           // List of server configurations
	Bot     struct {
		Token          string        `yaml:"token"`                         // Discord bot token
		UpdateInterval time.Duration `yaml:"update_interval" default:"30s"` // Interval for status updates
		Concurrency    int           `yaml:"concurrency" default:"10"`      // Number of concurrent operations
	} `yaml:"bot"`

	prevCumulativeOnline int // Internal state to track previous cumulative online count
}

/*
ServerConfig represents the configuration for an individual server.

It includes server connection details and Discord channel/category settings.
*/
type ServerConfig struct {
	// Configuration data

	ID           string `yaml:"id"`                            // Unique identifier for the server
	Host         string `yaml:"host" default:"127.0.0.1"`      // Server host address
	ChannelID    string `yaml:"channel_id,omitempty"`          // Discord channel ID to update
	ChannelName  string `yaml:"channel_name,omitempty"`        // Template for channel name
	ChannelDesc  string `yaml:"channel_description,omitempty"` // Template for channel description
	CategoryID   string `yaml:"category_id,omitempty"`         // Discord category ID to update
	CategoryName string `yaml:"category_name,omitempty"`       // Template for category name
	Port         int    `yaml:"port" default:"27016"`          // Server port
	Timeout      int    `yaml:"timeout" default:"3"`           // Timeout in seconds for A2S queries

	// Fields to store the previous state hashes for channels and categories

	prevChannelHash  uint64 // Previous hash for the channel
	prevCategoryHash uint64 // Previous hash for the category

	// Configuration data again (aligned)

	BufferSize uint16 `yaml:"buffer_size" default:"1024"` // Buffer size for A2S queries
}

/*
readConfig reads and parses the configuration file.

It loads the YAML configuration from the specified path,
applies default values, and sets up logging.

The function returns a pointer to a Config struct and an error if the operation fails.
*/
func readConfig() (*Config, error) {
	var path = "config.yaml"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if cfg.Bot.Token == "" {
		return nil, fmt.Errorf("bot token is empty")
	}

	defaults.SetDefaults(&cfg)
	cfg.Logging.setup()

	return &cfg, nil
}
