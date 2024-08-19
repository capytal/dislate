package bot

import "dislate/internals/discord/bot/events"

func (b *Bot) registerEventHandlers() {
	ehs := []any{
		events.NewGuildCreate(b.logger, b.db).Serve,
		events.NewMessageCreate(b.logger, b.db).Serve,
		events.NewReady(b.logger, b.db).Serve,
	}
	for _, h := range ehs {
		b.session.AddHandler(h)
	}
}