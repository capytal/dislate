package events

import (
	"dislate/internals/discord/bot/errors"
	"dislate/internals/discord/bot/gconf"
	e "errors"
	"log/slog"

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

func (h GuildCreate) Serve(s *dgo.Session, ev *dgo.GuildCreate) {
	err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: ev.Guild.ID})

	evErr := errors.NewEventError[GuildCreate](map[string]any{
		"GuildID": ev.Guild.ID,
	})

	if err != nil && !e.Is(err, gdb.ErrNoAffect) {
		evErr.Wrapf("Failed to add guild to database", err).Log(h.log)
	} else if err != nil {
		h.log.Info("Guild already in database", slog.String("id", ev.Guild.ID))
	} else {
		h.log.Info("Added guild", slog.String("id", ev.Guild.ID))
	}
}

type Ready struct {
	log *slog.Logger
	db  gconf.DB
}

func NewReady(log *slog.Logger, db gconf.DB) EventHandler[*dgo.Ready] {
	return Ready{log, db}
}

func (h Ready) Serve(s *dgo.Session, ev *dgo.Ready) {
	evErr := errors.NewEventError[GuildCreate](map[string]any{})

	for _, g := range ev.Guilds {
		err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: g.ID})

		if err != nil && !e.Is(err, gdb.ErrNoAffect) {
			evErr.Wrapf("Failed to add guild to database", err).AddData("GuildID", g.ID).Log(h.log)
		} else if err != nil {
			h.log.Info("Guild already in database", slog.String("id", g.ID))
		} else {
			h.log.Info("Added guild", slog.String("id", g.ID))
		}
	}
}
