package errors

import (
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
)

type ThreadCreateErr struct {
	*defaultEventErr[*dgo.ThreadCreate]
}

func NewThreadCreateErr(s *dgo.Session, ev *dgo.ThreadCreate, log *slog.Logger) ThreadCreateErr {
	return ThreadCreateErr{&defaultEventErr[*dgo.ThreadCreate]{
		data: map[string]any{
			"ThreadID": ev.ID,
			"ParentID": ev.ParentID,
			"GuildID":  ev.GuildID,
		},
		session:   s,
		channelID: ev.ID,
		logger:    log,
	}}
}
