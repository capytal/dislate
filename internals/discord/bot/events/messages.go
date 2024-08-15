package events

import (
	"log/slog"

	"dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type EventHandler[E any] interface {
	Serve(*dgo.Session, E)
}

type MessageCreate struct {
	log *slog.Logger
	db  guilddb.GuildDB
}

func NewMessageCreate(log *slog.Logger, db guilddb.GuildDB) MessageCreate {
	return MessageCreate{log, db}
}
func (h MessageCreate) Serve(s *dgo.Session, e *dgo.MessageCreate) {
	
}
