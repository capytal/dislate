package events

import (
	e "errors"
	"log/slog"

	"forge.capytal.company/capytal/dislate/bot/events/errors"
	"forge.capytal.company/capytal/dislate/bot/gconf"

	gdb "forge.capytal.company/capytal/dislate/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type GuildCreate struct {
	log *slog.Logger
	db  gconf.DB
}

func NewGuildCreate(log *slog.Logger, db gconf.DB) GuildCreate {
	return GuildCreate{log, db}
}

func (h GuildCreate) Serve(s *dgo.Session, ev *dgo.GuildCreate) errors.EventErr {
	err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: ev.Guild.ID})

	everr := errors.NewGuildErr[*dgo.GuildCreate](ev.Guild, h.log)

	if err != nil && !e.Is(err, gdb.ErrNoAffect) {
		return everr.Join(e.New("Failed to add guild to database"), err)
	} else if err != nil {
		h.log.Info("Guild already in database", slog.String("id", ev.Guild.ID))
	} else {
		h.log.Info("Added guild", slog.String("id", ev.Guild.ID))
	}

	return nil
}

type Ready struct {
	log *slog.Logger
	db  gconf.DB
}

func NewReady(log *slog.Logger, db gconf.DB) EventHandler[*dgo.Ready] {
	return Ready{log, db}
}

func (h Ready) Serve(s *dgo.Session, ev *dgo.Ready) errors.EventErr {
	everr := errors.NewReadyErr(ev, h.log)

	for _, g := range ev.Guilds {
		err := h.db.GuildInsert(gdb.Guild[gconf.ConfigString]{ID: g.ID})

		if err != nil && !e.Is(err, gdb.ErrNoAffect) {
			return everr.Join(err)
		} else if err != nil {
			h.log.Info("Guild already in database", slog.String("id", g.ID))
		} else {
			h.log.Info("Added guild", slog.String("id", g.ID))
		}
	}

	return nil
}
