package events

import (
	"dislate/internals/discord/bot/events/errors"

	dgo "github.com/bwmarrin/discordgo"
)

type EventHandler[E any] interface {
	Serve(*dgo.Session, E) errors.EventErr
}
