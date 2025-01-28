package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

/*
PresenceStats holds the statistics used to update Discord Rich Presence.

It includes the total number of servers, online servers,
current players, maximum slots, and players in the queue.
*/
type PresenceStats struct {
	Servers       int // Total number of servers
	OnlineServers int // Number of online servers
	Players       int // Current number of players
	Slots         int // Maximum number of player slots
	Queue         int // Number of players in the queue
}

/*
update updates the Discord Rich Presence based on the current statistics.

It checks if the cumulative online status has changed to avoid unnecessary updates.
If changes are detected, it updates the Rich Presence and logs the action.

Parameters:
  - ds: Discord session.
  - cfg: Configuration settings.

Returns an error if the update operation fails.
*/
func (p *PresenceStats) update(ds *discordgo.Session, cfg *Config) error {
	cumulative := p.OnlineServers + p.Players + p.Queue

	if cumulative == cfg.prevCumulativeOnline {
		log.Debug().Msg("Skipping Rich Presence update; no changes detected")
		return nil
	}

	if err := ds.UpdateStatusComplex(p.makeUSD()); err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}

	log.Debug().Msg("Rich Presence updated successfully")
	cfg.prevCumulativeOnline = cumulative

	return nil
}

/*
makeUSD creates the UpdateStatusData for Discord Rich Presence.

It constructs the presence message based on the current statistics,
ensuring it does not exceed Discord's character limit.
*/
func (p *PresenceStats) makeUSD() discordgo.UpdateStatusData {
	var presence, status string

	if p.OnlineServers == 0 {
		status = "idle"

		if p.Servers == 1 {
			presence = "Server offline"
		} else {
			presence = "All servers offline"
		}
	} else {
		status = "online"

		if p.Queue == 0 {
			presence = fmt.Sprintf("%d/%d players", p.Players, p.Slots)
		} else {
			presence = fmt.Sprintf("%d/%d (+%d) players", p.Players, p.Slots, p.Queue)
		}

		if p.Servers > 1 && p.OnlineServers > 0 {
			if p.OnlineServers < p.Servers {
				presence += fmt.Sprintf(" on %d/%d servers", p.OnlineServers, p.Servers)
			} else {
				presence += fmt.Sprintf(" on %d servers", p.OnlineServers)
			}
		}
	}

	if len(presence) > 128 {
		presence = presence[:125] + "..."
	}

	return discordgo.UpdateStatusData{
		Status: status,
		Activities: []*discordgo.Activity{
			{
				Name:  presence,
				Type:  discordgo.ActivityTypeCustom,
				State: presence,
			},
		},
	}
}
