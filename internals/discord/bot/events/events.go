package events

import (
	dgo "github.com/bwmarrin/discordgo"
)

type EventHandler[E any] interface {
	Serve(*dgo.Session, E)
}
