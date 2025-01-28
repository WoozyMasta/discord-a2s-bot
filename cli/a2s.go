package main

import (
	"github.com/rs/zerolog/log"
	"github.com/woozymasta/a2s/pkg/a2s"
)

/*
getInfo queries the A2S server and returns the server information.

It creates a new A2S client with the server's host and port,
sets the buffer size and timeout as per the ServerConfig,
and retrieves the server information.

Returns a pointer to a2s.Info containing server details and an error if the operation fails.
*/
func (s *ServerConfig) getInfo() (*a2s.Info, error) {
	client, err := a2s.New(s.Host, s.Port)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Error().Err(err).Msg("Error close A2S client")
		}
	}()

	client.SetBufferSize(s.BufferSize)
	client.SetDeadlineTimeout(s.Timeout)

	return client.GetInfo()
}
