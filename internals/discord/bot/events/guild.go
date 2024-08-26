package events

import (
	"errors"
	"log/slog"

	"dislate/internals/discord/bot/gconf"
	gdb "dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type GuildCreate struct {
	log *slog.Logger
	db  gconf.DB
}

func NewGuildCreate(log *slog.Logger, db gconf.DB) GuildCreate {
	return GuildCreate{log, db}
}
func (h GuildCreate) Serve(s *dgo.Session, e *dgo.GuildCreate) {
	err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: e.Guild.ID})

	if err != nil && !errors.Is(err, gdb.ErrNoAffect) {
		h.log.Error("Failed to add guild to database",
			slog.String("id", e.Guild.ID),
			slog.String("err", err.Error()),
		)
	} else if err != nil {
		h.log.Info("Guild already in database", slog.String("id", e.Guild.ID))
	} else {
		h.log.Info("Added guild", slog.String("id", e.Guild.ID))
	}
}

type Ready struct {
	log *slog.Logger
	db  gconf.DB
}

func NewReady(log *slog.Logger, db gconf.DB) EventHandler[*dgo.Ready] {
	return Ready{log, db}
}
func (h Ready) Serve(s *dgo.Session, e *dgo.Ready) {
	for _, g := range e.Guilds {
		err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: g.ID})

		if err != nil && !errors.Is(err, gdb.ErrNoAffect) {
			h.log.Error("Failed to add guild to database",
				slog.String("id", g.ID),
				slog.String("err", err.Error()),
			)
		} else if err != nil {
			h.log.Info("Guild already in database", slog.String("id", g.ID))
		} else {
			h.log.Info("Added guild", slog.String("id", g.ID))
		}
	}
}
