// update.go

package main

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/woozymasta/a2s/pkg/keywords"
	"github.com/woozymasta/steam/utils/appid"
)

/*
update performs the Rich Presence update and enqueues channel updates.

Steps:
 1. Query each server in parallel to get info.
 2. Update aggregated stats for Rich Presence.
 3. Immediately update Rich Presence (fast).
 4. Enqueue tasks to update channels/categories (async).
*/
func update(ds *discordgo.Session, cfg *Config) {
	stats := &PresenceStats{Servers: len(cfg.Servers)}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.Bot.Concurrency)

	for i := range cfg.Servers {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			srv := &cfg.Servers[i]
			tplData := &TemplateData{
				ID:   srv.ID,
				Host: srv.Host,
				Port: srv.Port,
			}

			log.Debug().
				Str("server", srv.ID).
				Str("host", fmt.Sprintf("%s:%d", srv.Host, srv.Port)).
				Msg("Querying server")

			info, err := srv.getInfo()
			if err != nil {
				log.Warn().Err(err).Str("server", srv.ID).Msgf("Failed to retrieve information for server")
				// If server is offline, we still might want to update channel to "offline".
				// Enqueue with nil Info
				channelUpdateQueue <- ChannelUpdateTask{Server: srv, Tpl: tplData}
				return
			}
			tplData.Info = info

			// Parse extra keywords if needed
			localQueue := 0
			switch info.ID {
			case appid.Arma3.Uint64():
				armaInfo := keywords.ParseArma3(info.Keywords)
				tplData.Extra = armaInfo

			case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
				dayzInfo := keywords.ParseDayZ(info.Keywords)
				localQueue = int(dayzInfo.PlayersQueue)
				tplData.Extra = dayzInfo
			}

			// Aggregate stats
			mu.Lock()
			stats.Players += int(info.Players)
			stats.Slots += int(info.MaxPlayers)
			stats.Queue += localQueue
			stats.OnlineServers++
			mu.Unlock()

			// Enqueue the update of channel/category asynchronously
			channelUpdateQueue <- ChannelUpdateTask{Server: srv, Tpl: tplData}
		}(i)
	}

	wg.Wait()

	// Update Discord Rich Presence with aggregated stats (immediate)
	if err := stats.update(ds, cfg); err != nil {
		log.Error().Err(err).Msg("Error updating Rich Presence")
	}

	log.Info().Msg("Update completed")
}
