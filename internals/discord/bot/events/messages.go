package events

import (
	e "errors"
	"log/slog"
	"slices"

	"dislate/internals/discord/bot/errors"
	"dislate/internals/discord/bot/gconf"
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"dislate/internals/translator/lang"

	dgo "github.com/bwmarrin/discordgo"
)

type MessageCreate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewMessageCreate(db gconf.DB, t translator.Translator) MessageCreate {
	return MessageCreate{db, t}
}

func (h MessageCreate) Serve(s *dgo.Session, ev *dgo.MessageCreate) {
	if ev.Message.Author.Bot {
		return
	}
	log := gconf.GetLogger(ev.GuildID, s, h.db)

	ch, err := h.db.Channel(ev.GuildID, ev.ChannelID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in database, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return
	} else if err != nil {
		errors.NewErrDatabase(
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
			slog.String("err", err.Error()),
		).LogReply(log, s, ev.Message)
		return
	}

	gc, err := h.db.ChannelGroup(ch.GuildID, ch.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in a group, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return
	} else if err != nil {
		errors.NewErrDatabase(
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
			slog.String("err", err.Error()),
		).LogReply(log, s, ev.Message)
		return
	}

	_, err = getMessage(h.db, ev.Message, ch.Language)
	if err != nil {
		errors.NewErrDatabase(
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
			slog.String("message", ev.Message.ID),
			slog.String("err", err.Error()),
		).LogReply(log, s, ev.Message)
		return
	}

	for _, c := range gc {
		if c.ID == ch.ID && c.GuildID == ch.GuildID {
			continue
		}
		go func(c guilddb.Channel) {
			uw, err := getUserWebhook(s, c.ID, ev.Message.Author)
			if err != nil {
				errors.NewErrUserWebhook(
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.Any("user", ev.Message.Author),
				).LogReply(log, s, ev.Message)
			}

			t, err := h.translator.Translate(ch.Language, c.Language, ev.Message.Content)
			if err != nil {
				errors.New("Error while trying to translate message",
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.String("message", ev.Message.ID),
					slog.String("content", ev.Message.Content),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
			}

			tdm, err := s.WebhookExecute(uw.ID, uw.Token, true, &dgo.WebhookParams{
				AvatarURL: ev.Message.Author.AvatarURL(""),
				Username:  ev.Message.Author.GlobalName,
				Content:   t,
			})
			if err != nil {
				errors.NewErrUserWebhook(
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.String("message", ev.Message.ID),
					slog.String("content", ev.Message.Content),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
			}

			if tdm.GuildID == "" {
				tdm.GuildID = ev.Message.GuildID
			}

			_, err = getTranslatedMessage(h.db, tdm, ev.Message, c.Language)
			if err != nil {
				errors.NewErrDatabase(
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.String("message", ev.Message.ID),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
			}
		}(c)

	}
}

type MessageEdit struct {
	db         gconf.DB
	translator translator.Translator
}

func NewMessageEdit(db gconf.DB, t translator.Translator) MessageEdit {
	return MessageEdit{db, t}
}

func (h MessageEdit) Serve(s *dgo.Session, ev *dgo.MessageUpdate) {
	if ev.Message.Author.Bot {
		return
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)

	msg, err := getMessage(h.db, ev.Message, lang.EN)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Message is not in database, ignoring.",
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
		)
		return
	} else if err != nil {
		errors.NewErrDatabase(
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
			slog.String("err", err.Error()),
		).LogReply(log, s, ev.Message)
		return
	}

	tmsgs, err := h.db.MessagesWithOrigin(msg.GuildID, msg.ChannelID, msg.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("No translated message found, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return
	} else if err != nil {
		errors.NewErrDatabase(
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
			slog.String("err", err.Error()),
		).LogReply(log, s, ev.Message)
		return
	}

	for _, m := range tmsgs {
		if m.ID == msg.ID && m.GuildID == msg.GuildID {
			continue
		}
		go func(m guilddb.Message) {
			uw, err := getUserWebhook(s, m.ChannelID, ev.Message.Author)
			if err != nil {
				errors.NewErrUserWebhook(
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.Any("user", ev.Message.Author),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
				return
			}

			t, err := h.translator.Translate(msg.Language, m.Language, ev.Message.Content)
			if err != nil {
				errors.New("Error while trying to translate message",
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.String("message", ev.Message.ID),
					slog.String("content", ev.Message.Content),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
				return
			}

			_, err = s.WebhookMessageEdit(uw.ID, uw.Token, m.ID, &dgo.WebhookEdit{
				Content: &t,
			})
			if err != nil {
				errors.NewErrUserWebhook(
					slog.String("guild", ev.Message.GuildID),
					slog.String("channel", ev.Message.ChannelID),
					slog.String("message", ev.Message.ID),
					slog.String("content", ev.Message.Content),
					slog.String("err", err.Error()),
				).LogReply(log, s, ev.Message)
				return
			}
		}(m)

	}
}

func getUserWebhook(s *dgo.Session, channelID string, user *dgo.User) (*dgo.Webhook, error) {
	whName := "DISLATE_USER_WEBHOOK_" + user.ID

	ws, err := s.ChannelWebhooks(channelID)
	if err != nil {
		return &dgo.Webhook{}, err
	}
	wi := slices.IndexFunc(ws, func(w *dgo.Webhook) bool {
		return w.Name == whName
	})

	if wi > -1 {
		return ws[wi], nil
	}

	w, err := s.WebhookCreate(channelID, whName, user.AvatarURL(""))
	if err != nil {
		return &dgo.Webhook{}, err
	}

	return w, nil
}

func getMessage(db gconf.DB, m *dgo.Message, lang lang.Language) (guilddb.Message, error) {
	msg, err := db.Message(m.GuildID, m.ChannelID, m.ID)

	if e.Is(err, guilddb.ErrNotFound) {
		if err := db.MessageInsert(guilddb.NewMessage(m.GuildID, m.ChannelID, m.ID, lang)); err != nil {
			return guilddb.Message{}, err
		}
		msg, err = db.Message(m.GuildID, m.ChannelID, m.ID)
		if err != nil {
			return guilddb.Message{}, err
		}
	}

	return msg, nil
}

func getTranslatedMessage(
	db gconf.DB,
	m, original *dgo.Message,
	lang lang.Language,
) (guilddb.Message, error) {
	msg, err := db.Message(m.GuildID, m.ChannelID, m.ID)

	if e.Is(err, guilddb.ErrNotFound) {
		if err := db.MessageInsert(guilddb.NewTranslatedMessage(
			m.GuildID,
			m.ChannelID,
			m.ID,
			lang,
			original.ChannelID,
			original.ID,
		)); err != nil {
			return guilddb.Message{}, err
		}
		msg, err = db.Message(m.GuildID, m.ChannelID, m.ID)
		if err != nil {
			return guilddb.Message{}, err
		}
	} else if err != nil {
		return guilddb.Message{}, err
	}

	return msg, nil
}
