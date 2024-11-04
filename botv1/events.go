package bot

import (
	"forge.capytal.company/capytal/dislate/bot/events"

	dgo "github.com/bwmarrin/discordgo"
)

func w[E any](h events.EventHandler[E]) interface{} {
	return func(s *dgo.Session, ev E) {
		err := h.Serve(s, ev)
		if err != nil {
			err.Log()
			err.Send()
			err.Reply()
		}
	}
}

func (b *Bot) registerEventHandlers() {
	ehs := []any{
		w(events.NewGuildCreate(b.logger, b.db)),
		w(events.NewMessageCreate(b.db, b.translator)),
		w(events.NewMessageUpdate(b.db, b.translator)),
		w(events.NewMessageDelete(b.db)),
		w(events.NewReady(b.logger, b.db)),
		w(events.NewThreadCreate(b.db, b.translator)),
	}
	for _, h := range ehs {
		b.session.AddHandler(h)
	}
}
