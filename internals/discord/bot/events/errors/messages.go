package errors

import (
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
)

type MessageErr[E any] struct {
	*defaultEventErr[E]
}

func NewMessageErr[E any](
	s *dgo.Session,
	msg *dgo.Message,
	log *slog.Logger,
) MessageErr[E] {
	return MessageErr[E]{&defaultEventErr[E]{
		data: map[string]any{
			"MessageID": msg.ID,
			"ChannelID": msg.ChannelID,
			"GuildID":   msg.GuildID,
			"AuthorID":  msg.Author.ID,
		},
		session:          s,
		channelID:        msg.ChannelID,
		messageReference: msg.Reference(),
		logger:           log,
	}}
}
