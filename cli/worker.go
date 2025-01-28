package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

/*
ChannelUpdateTask describes a single update operation for a server:
it will update channel and category (if needed).
*/
type ChannelUpdateTask struct {
	Server *ServerConfig
	Tpl    *TemplateData
}

// channelUpdateQueue is a buffered channel to store update tasks
var channelUpdateQueue = make(chan ChannelUpdateTask, 100)

/*
startUpdateWorkers launches 'workerCount' goroutines that read tasks from channelUpdateQueue
and process them. Each task updates the server's channel/category in a blocking call, but
this doesn't block the main update() because it's done asynchronously.
*/
func startUpdateWorkers(ds *discordgo.Session, workerCount int, timeout time.Duration) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for task := range channelUpdateQueue {
				processChannelUpdate(ds, task, timeout)
			}
		}()
	}
}

/*
processChannelUpdate handles channel and category updates for one server config.
It uses a context with timeout to avoid being stuck if Discord is slow.
*/
func processChannelUpdate(ds *discordgo.Session, task ChannelUpdateTask, timeout time.Duration) {
	if ds == nil || task.Server == nil || task.Tpl == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Update channel
	if err := task.Server.updateChannel(ctx, ds, task.Tpl); err != nil {
		log.Error().
			Err(err).
			Str("server", task.Server.ID).
			Msg("Failed to update channel for server")
	}

	// Update category
	if err := task.Server.updateCategory(ctx, ds, task.Tpl); err != nil {
		log.Error().
			Err(err).
			Str("server", task.Server.ID).
			Msgf("Failed to update category for server")
	}
}

/*
editChannel is a context-aware function that edits the channel's name/topic.

By default, discordgo doesn't have ChannelEditContext,
so we manually check ctx.Done() before and after the request.
You could also replace the http.Client in the session if you want real cancellation.
*/
func editChannel(ctx context.Context, ds *discordgo.Session, id, name, description string) error {
	if id == "" {
		return nil
	}

	if name == "" && description == "" {
		log.Debug().
			Str("id", id).
			Msg("Nothing to edit in channel/category")
		return nil
	}

	if len(name) > 25 {
		log.Warn().
			Str("name", name).
			Msg("Channel name more then 25 characters, its not recommended length")
	}

	log.Debug().
		Str("id", id).
		Str("name", name).
		Str("description", description).
		Msg("Preparing to edit channel/category")

	select {
	case <-ctx.Done():
		return fmt.Errorf("edit channel canceled before request")
	default:
	}

	ce := &discordgo.ChannelEdit{}
	if name != "" {
		ce.Name = name
	}
	if description != "" {
		ce.Topic = description
	}

	_, err := ds.ChannelEdit(id, ce)

	select {
	case <-ctx.Done():
		return fmt.Errorf("edit channel canceled after request")
	default:
	}

	return err
}
