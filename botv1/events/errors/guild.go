package errors

import (
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
)

type GuildErr[E any] struct {
	*defaultEventErr[E]
}

func NewGuildErr[E any](
	g *dgo.Guild,
	log *slog.Logger,
) GuildErr[E] {
	return GuildErr[E]{&defaultEventErr[E]{
		data: map[string]any{
			"GuildID": g.ID,
		},
		logger: log,
	}}
}

type ReadyErr struct {
	*defaultEventErr[*dgo.Ready]
}

func NewReadyErr(
	ev *dgo.Ready,
	log *slog.Logger,
) ReadyErr {
	return ReadyErr{&defaultEventErr[*dgo.Ready]{
		data: map[string]any{
			"SessionID":   ev.SessionID,
			"BotUserID":   ev.User.ID,
			"BotUserName": ev.User.Username,
			"Guilds":      ev.Guilds,
		},
		logger: log,
	}}
}
