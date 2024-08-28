package bot

import "dislate/internals/discord/bot/events"

func (b *Bot) registerEventHandlers() {
	ehs := []any{
		events.NewGuildCreate(b.logger, b.db).Serve,
		events.NewMessageCreate(b.db, b.translator).Serve,
		events.NewMessageEdit(b.db, b.translator).Serve,
		events.NewReady(b.logger, b.db).Serve,
		events.NewThreadCreate(b.db, b.translator).Serve,
	}
	for _, h := range ehs {
		b.session.AddHandler(h)
	}
}
