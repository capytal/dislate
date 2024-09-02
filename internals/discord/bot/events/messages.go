package events

import (
	"dislate/internals/discord/bot/errors"
	"dislate/internals/discord/bot/gconf"
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"dislate/internals/translator/lang"
	e "errors"
	"log/slog"
	"slices"

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
	if ev.Message.Author.Bot || ev.Type != dgo.MessageTypeDefault {
		return
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)
	h.sendMessage(log, s, ev.Message)
}

func (h MessageCreate) sendMessage(log *slog.Logger, s *dgo.Session, msg *dgo.Message) {
	everr := errors.NewEventError[MessageCreate](map[string]any{
		"GuildID":   msg.GuildID,
		"ChannelID": msg.ChannelID,
		"MessageID": msg.ID,
	})

	ch, err := h.db.Channel(msg.GuildID, msg.ChannelID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in database, ignoring.",
			slog.String("guild", msg.GuildID),
			slog.String("channel", msg.ChannelID),
			slog.String("message", msg.ID),
		)
		return
	} else if err != nil {
		everr.Wrapf("Failed to get channel from database", err).Log(log).Reply(s, msg)
		return
	}

	gc, err := h.db.ChannelGroup(ch.GuildID, ch.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in a group, ignoring.",
			slog.String("guild", msg.GuildID),
			slog.String("channel", msg.ChannelID),
			slog.String("message", msg.ID),
		)
		return
	} else if err != nil {
		everr.Wrapf("Failed to get channel group from database", err).Log(log).Reply(s, msg)
		return
	}

	_, err = getMessage(h.db, msg, ch.Language)
	if err != nil {
		everr.Wrapf("Failed to get/add message to database", err).Log(log).Reply(s, msg)
		return
	}

	for _, c := range gc {
		if c.ID == ch.ID && c.GuildID == ch.GuildID {
			continue
		}
		go func(c guilddb.Channel) {
			dch, err := s.Channel(c.ID)

			var channelID string
			if err != nil {
				everr.Wrapf("Failed to get information about translated channel", err).
					AddData("TranslatedChannel", c.ID).
					Log(log).
					Reply(s, msg)
				return
			} else if dch.IsThread() {
				channelID = dch.ParentID
			} else {
				channelID = dch.ID
			}

			uw, err := getUserWebhook(s, channelID, msg.Author)
			if err != nil {
				everr.Wrapf("Failed to get/set user webhook for translated channel", err).
					AddData("TranslatedChannel", c.ID).
					AddData("User", msg.Author.ID).
					Log(log).
					Reply(s, msg)
				return
			}

			t, err := h.translator.Translate(ch.Language, c.Language, msg.Content)
			if err != nil {
				everr.Wrapf("Error while trying to translate message", err).
					AddData("content", msg.Content).
					Log(log).
					Reply(s, msg)
				return
			}

			var tdm *dgo.Message
			if dch.IsThread() {
				tdm, err = s.WebhookThreadExecute(uw.ID, uw.Token, true, dch.ID, &dgo.WebhookParams{
					AvatarURL: msg.Author.AvatarURL(""),
					Username:  msg.Author.GlobalName,
					Content:   t,
				})
			} else {
				tdm, err = s.WebhookExecute(uw.ID, uw.Token, true, &dgo.WebhookParams{
					AvatarURL: msg.Author.AvatarURL(""),
					Username:  msg.Author.GlobalName,
					Content:   t,
				})
			}
			if err != nil {
				everr.Wrapf("Error while trying to execute user webhook", err).
					AddData("content", msg.Content).
					AddData("User", msg.Author.ID).
					AddData("Webhook", uw.ID).
					Log(log).
					Reply(s, msg)
				return
			}

			if tdm.GuildID == "" {
				tdm.GuildID = msg.GuildID
			}

			_, err = getTranslatedMessage(h.db, tdm, msg, c.Language)
			if err != nil {
				everr.Wrapf("Error while trying to get/set translated message", err).
					AddData("TranslatedMessageID", tdm.ID).
					Log(log).
					Reply(s, msg)
				return
			}
		}(c)

	}
}

type MessageUpdate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewMessageUpdate(db gconf.DB, t translator.Translator) MessageUpdate {
	return MessageUpdate{db, t}
}

func (h MessageUpdate) Serve(s *dgo.Session, ev *dgo.MessageUpdate) {
	if ev.Message.Author.Bot || ev.Type != dgo.MessageTypeDefault {
		return
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)

	everr := errors.NewEventError[MessageUpdate](map[string]any{
		"GuildID":   ev.Message.GuildID,
		"ChannelID": ev.Message.ChannelID,
		"MessageID": ev.Message.ID,
	})

	msg, err := h.db.Message(ev.Message.GuildID, ev.Message.ChannelID, ev.Message.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Message is not in database, ignoring.",
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
		)
		return
	} else if err != nil {
		everr.Wrapf("Failed to get message from database", err).Log(log).Reply(s, ev.Message)
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
		everr.Wrapf("Failed to get translated messages from database", err).Log(log).Reply(s, ev.Message)
		return
	}

	for _, m := range tmsgs {
		if m.ID == msg.ID && m.GuildID == msg.GuildID {
			continue
		}
		go func(m guilddb.Message) {
			var channelID string
			if dch, err := s.Channel(m.ChannelID); err != nil {
				everr.Wrapf("Failed to get information about translated channel", err).
					AddData("TranslatedChannel", m.ChannelID).
					Log(log).
					Reply(s, ev.Message)
			} else if dch.IsThread() {
				channelID = dch.ParentID
			} else {
				channelID = dch.ID
			}

			uw, err := getUserWebhook(s, channelID, ev.Message.Author)
			if err != nil {
				everr.Wrapf("Failed to get/set user webhook for translated channel", err).
					AddData("TranslatedChannel", m.ChannelID).
					AddData("User", ev.Message.Author.ID).
					Log(log).
					Reply(s, ev.Message)
				return
			}

			t, err := h.translator.Translate(msg.Language, m.Language, ev.Message.Content)
			if err != nil {
				everr.Wrapf("Error while trying to translate message", err).
					AddData("content", ev.Message.Content).
					Log(log).
					Reply(s, ev.Message)
				return
			}

			_, err = s.WebhookMessageEdit(uw.ID, uw.Token, m.ID, &dgo.WebhookEdit{
				Content: &t,
			})
			if err != nil {
				everr.Wrapf("Error while trying to execute user webhook", err).
					AddData("content", ev.Message.Content).
					AddData("User", ev.Message.Author.ID).
					AddData("Webhook", uw.ID).
					Log(log).
					Reply(s, ev.Message)
				return
			}
		}(m)

	}
}

type MessageDelete struct {
	db gconf.DB
}

func NewMessageDelete(db gconf.DB) MessageDelete {
	return MessageDelete{db}
}

func (h MessageDelete) Serve(s *dgo.Session, ev *dgo.MessageDelete) {
	if ev.Type != dgo.MessageTypeDefault {
		return
	}
	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)

	everr := errors.NewEventError[MessageUpdate](map[string]any{
		"GuildID":   ev.Message.GuildID,
		"ChannelID": ev.Message.ChannelID,
		"MessageID": ev.Message.ID,
	})

	msg, err := h.db.Message(ev.Message.GuildID, ev.Message.ChannelID, ev.Message.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Message is not in database, ignoring.",
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
		)
		return
	} else if err != nil {
		everr.Wrapf("Failed to get message from database", err).Log(log).Reply(s, ev.Message)
		return
	}

	var originChannelID, originID string
	if msg.OriginID != nil && msg.OriginChannelID != nil {
		oMsg, err := h.db.Message(ev.Message.GuildID, *msg.OriginChannelID, *msg.OriginID)
		if err != nil {
			originChannelID, originID = *msg.OriginChannelID, *msg.OriginID
		} else {
			msg, originChannelID, originID = oMsg, oMsg.ChannelID, oMsg.ID
		}
	} else {
		originChannelID, originID = msg.ChannelID, msg.ID
	}

	tmsgs, err := h.db.MessagesWithOrigin(msg.GuildID, originChannelID, originID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("No translated message found, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return
	} else if err != nil {
		everr.Wrapf("Failed to get translated messages from database", err).Log(log).Reply(s, ev.Message)
		return
	}

	for _, m := range tmsgs {
		if m.ID == msg.ID && m.ChannelID == msg.ChannelID && m.GuildID == msg.GuildID {
			continue
		}
		go func(m guilddb.Message) {
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				log.Warn("Failed to delete message",
					slog.String("channel", m.ChannelID),
					slog.String("message", m.ID),
					slog.String("err", err.Error()),
				)
			}
		}(m)
	}

	if err := s.ChannelMessageDelete(msg.ChannelID, msg.ID); err != nil {
		log.Warn("Failed to delete message",
			slog.String("channel", msg.ChannelID),
			slog.String("message", msg.ID),
			slog.String("err", err.Error()),
		)
	}

	if err := h.db.MessageDelete(guilddb.NewMessage(msg.GuildID, msg.ChannelID, msg.ID, lang.EN)); err != nil {
		everr.Wrapf("Failed to delete message from database", err).Log(log).Send(s, msg.ChannelID)
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
