// channel.go

package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/zeebo/xxh3"
)

// updateChannel attempts to render the channel's template, compare hash, edit if needed
func (s *ServerConfig) updateChannel(ctx context.Context, ds *discordgo.Session, tpl *TemplateData) error {
	if s.ChannelID == "" || ds == nil || tpl == nil {
		return nil
	}
	if s.ChannelName == "" && s.ChannelDesc == "" {
		return nil
	}

	// Render templates
	var name, description string
	if s.ChannelName != "" {
		rendered, err := tpl.render(s.ChannelName)
		if err != nil {
			log.Error().Err(err).Str("channel", s.ChannelID).Msg("Error rendering channel name template")
		}
		name = rendered
	}

	if s.ChannelDesc != "" {
		rendered, err := tpl.render(s.ChannelDesc)
		if err != nil {
			log.Error().Err(err).Str("channel", s.ChannelID).Msg("Error rendering channel description template")
		} else {
			description = rendered

			if len(description) > 1024 {
				description = description[:1021] + "..."
			}
		}
	}

	// Compare hashes
	newHash := xxh3.HashString(name + description)
	if newHash == s.prevChannelHash {
		log.Debug().Str("channel", s.ChannelID).Msg("Skipping update for channel without changes detected")
		return nil
	}

	log.Debug().
		Str("channel", s.ChannelID).
		Uint64("new hash", newHash).
		Uint64("prev hash", s.prevChannelHash).
		Msgf("Updating channel")

	// Actual edit (context-based)
	err := editChannel(ctx, ds, s.ChannelID, name, description)
	if err != nil {
		return err
	}

	// If successful
	s.prevChannelHash = newHash
	return nil
}

// updateCategory attempts to render the category's template, compare hash, edit if needed
func (s *ServerConfig) updateCategory(ctx context.Context, ds *discordgo.Session, tpl *TemplateData) error {
	if s.CategoryID == "" || s.CategoryName == "" || ds == nil || tpl == nil {
		return nil
	}

	name, err := tpl.render(s.CategoryName)
	if err != nil {
		log.Error().Err(err).Str("channel", s.CategoryID).Msg("Error rendering category name template")
		name = ""
	}

	newHash := xxh3.HashString(name)
	if newHash == s.prevCategoryHash {
		log.Debug().Str("channel", s.CategoryID).Msg("Skipping update for category without changes detected")
		return nil
	}

	log.Debug().
		Str("category", s.ChannelID).
		Uint64("new hash", newHash).
		Uint64("prev hash", s.prevCategoryHash).
		Msgf("Updating category")

	err = editChannel(ctx, ds, s.CategoryID, name, "")
	if err != nil {
		return err
	}
	s.prevCategoryHash = newHash

	return nil
}
