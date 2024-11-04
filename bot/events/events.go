package events

import (
	"forge.capytal.company/capytal/dislate/bot/events/errors"

	dgo "github.com/bwmarrin/discordgo"
)

type EventHandler[E any] interface {
	Serve(*dgo.Session, E) errors.EventErr
}
