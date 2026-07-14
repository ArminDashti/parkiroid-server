package livekit

import "github.com/dogan/dogan-server/internal/models"

func DefaultIceServers() []models.IceServerConfig {
	return []models.IceServerConfig{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}
}
