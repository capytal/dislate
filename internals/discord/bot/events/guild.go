package events

import (
	"errors"
	"log/slog"

	"dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type GuildCreate struct {
	log *slog.Logger
	db  guilddb.GuildDB
}

func NewGuildCreate(log *slog.Logger, db guilddb.GuildDB) GuildCreate {
	return GuildCreate{log, db}
}
func (h GuildCreate) Serve(s *dgo.Session, e *dgo.GuildCreate) {
	err := h.db.GuildInsert(guilddb.Guild{ID: e.Guild.ID})

	if err != nil && !errors.Is(err, guilddb.ErrNoAffect) {
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
	db  guilddb.GuildDB
}

func NewReady(log *slog.Logger, db guilddb.GuildDB) EventHandler[*dgo.Ready] {
	return Ready{log, db}
}
func (h Ready) Serve(s *dgo.Session, e *dgo.Ready) {
	for _, g := range e.Guilds {
		err := h.db.GuildInsert(guilddb.Guild{ID: g.ID})

		if err != nil && !errors.Is(err, guilddb.ErrNoAffect) {
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
