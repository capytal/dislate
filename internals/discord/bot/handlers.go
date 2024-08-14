package bot

import (
	"log"

	"dislate/internals/guilddb"

	"github.com/bwmarrin/discordgo"
)

type handler any

func (b *Bot) handlers() []handler {
	hs := []handler{
		func(s *discordgo.Session, r *discordgo.Ready) {
			for _, g := range r.Guilds {
				err := b.db.GuildInsert(guilddb.Guild{ID: g.ID})
				if err != nil {
					log.Printf("ERROR: Failed to add guild %s to database: %s", g.ID, err)
				} else {
					log.Printf("Added guild %s", g.ID)
				}
			}
		},
	}

	return hs
}

func (b *Bot) registerHandlers() {
	for _, h := range b.handlers() {
		b.session.AddHandler(h)
	}
}
