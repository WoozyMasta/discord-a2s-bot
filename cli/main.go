/*
Is a Discord bot that monitors game servers using A2S queries and updates Discord channels and
Rich Presence based on server statuses.

The bot periodically queries configured game servers, updates corresponding Discord channels with
server information, and maintains a Rich Presence status reflecting the overall server status.

Features:
  - Concurrently monitors multiple game servers.
  - Updates Discord channel names and descriptions based on server status.
  - Maintains Rich Presence with aggregated server statistics.
  - Supports customizable templates for channel and category names/descriptions.
  - Configurable logging with support for different log levels and outputs.

Usage:
  - Configure the bot using a YAML configuration file.
  - Run the bot, optionally specifying a configuration file path.
*/
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/woozymasta/discord-a2s-bot/internal/service"
)

func main() {
	if service.IsServiceMode() {
		service.RunAsService(runApp)
		return
	}

	runApp()
}

func runApp() {
	parseArgs()

	cfg, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading configuration")
	}

	// Create a new Discord session using the bot token from the configuration.
	dg, err := discordgo.New("Bot " + cfg.Bot.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Discord session")
	}

	// Channel to wait for the Ready event.
	ready := make(chan struct{})

	// Add a handler for the Ready event.
	dg.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		log.Debug().Msgf("Bot session %s opened", r.SessionID)
		close(ready)
	})

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening Discord session")
	}
	defer func() {
		if err := dg.Close(); err != nil {
			log.Error().Err(err).Msg("Error close Discord websocket connection")
		}
	}()

	log.Info().Msg("Bot connected to Discord")

	// Wait for the Ready event before proceeding.
	<-ready

	// --- Start worker pool for async channel/category updates ---
	// Use concurrency from config, and some timeout for blocking calls (e.g. 30s).
	startUpdateWorkers(dg, cfg.Bot.Concurrency, 30*time.Second)

	// Create a ticker that triggers at intervals specified in the configuration.
	ticker := time.NewTicker(cfg.Bot.UpdateInterval)
	defer ticker.Stop()

	// Channel to listen for OS interrupt signals (e.g., Ctrl+C, SIGTERM).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Perform an initial update before entering the update loop.
	update(dg, cfg)

	// WaitGroup to ensure all goroutines finish before exiting.
	var wg sync.WaitGroup
	wg.Add(1)

	// Main loop: every tick we call update(), or stop if SIGINT/SIGTERM
	for {
		select {
		case <-ticker.C:
			update(dg, cfg)
		case <-stop:
			// Received a termination signal, initiate shutdown.
			log.Info().Msg("Termination signal received. Stopping the bot...")
			log.Info().Msg("The bot has successfully shut down")
			return
		}
	}
}
