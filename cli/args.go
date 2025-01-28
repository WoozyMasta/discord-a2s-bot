package main

import (
	_ "embed"

	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/discord-a2s-bot/internal/vars"
)

//go:embed example.config.yaml
var exampleConfig string

// arguments parser
func parseArgs() {
	if len(os.Args) < 2 || !strings.HasPrefix(os.Args[1], "-") {
		return
	}

	switch os.Args[1] {
	case "--help", "-h":
		printHelp()
	case "--version", "-v":
		printVersion()
	case "--example", "-e":
		fmt.Println(exampleConfig)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command. Use --help for a list of available commands.")
		os.Exit(0)
	}
}

// just print help message and exit
func printHelp() {
	fmt.Printf(`%[1]s %s

Is a Discord bot that monitors game servers using A2S queries and updates Discord channels and Rich Presence based on server statuses.

Usage:
  %[1]s [option] [config.(yaml|json)]

Available options:
  -e, --example    Prints an example YAML configuration file.
  -v, --version    Show version, commit, and build time.
  -h, --help       Prints this help message.

Examples:
  Save an example YAML configuration to file:
    %[1]s -e > config.yaml

  Print an example environment variables:
    %[1]s -e | yq 'explode(.) | del(.base-template)' > config.yaml

`, filepath.Base(os.Args[0]), vars.Version)
	os.Exit(0)
}

// print version information message and exit
func printVersion() {
	fmt.Printf(`file:     %s
version:  %s
commit:   %s
built:    %s
project:  %s
`, os.Args[0], vars.Version, vars.Commit, vars.BuildTime, vars.URL)
	os.Exit(0)
}
