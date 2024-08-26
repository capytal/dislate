package gconf

import (
	"context"
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
)

type guildHandler struct {
	*slog.TextHandler
}

func NewGuildHandler(s *dgo.Session, c *dgo.Channel, opts *slog.HandlerOptions) guildHandler {
	w := NewChannelWriter(s, c)
	h := slog.NewTextHandler(w, opts)
	return guildHandler{h}
}

type disabledHandler struct {
	*slog.TextHandler
}
func (_ disabledHandler) Enabled(_ context.Context,_ slog.Level) bool {
	return false
}

type channelWriter struct {
	session *dgo.Session
	channel *dgo.Channel
}

func NewChannelWriter(s *dgo.Session, c *dgo.Channel) channelWriter {
	w := channelWriter{s, c}

	return w
}

func (w channelWriter) Write(p []byte) (int, error) {
	m, err := w.session.ChannelMessageSend(w.channel.ID, string(p))
	if err != nil {
		return 0, err
	}

	return len(m.Content), nil
}
