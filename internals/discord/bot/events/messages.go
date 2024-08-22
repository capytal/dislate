package events

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"dislate/internals/translator/lang"

	dgo "github.com/bwmarrin/discordgo"
)

type EventHandler[E any] interface {
	Serve(*dgo.Session, E)
}

type MessageCreate struct {
	log        *slog.Logger
	db         guilddb.GuildDB
	translator translator.Translator
}

func NewMessageCreate(log *slog.Logger, db guilddb.GuildDB, t translator.Translator) MessageCreate {
	return MessageCreate{log, db, t}
}
func (h MessageCreate) Serve(s *dgo.Session, e *dgo.MessageCreate) {
	if e.Message.Author.Bot {
		return
	}

	ch, err := h.db.Channel(e.GuildID, e.ChannelID)
	if errors.Is(err, guilddb.ErrNotFound) {
		h.log.Debug("Channel is not in database, ignoring.", slog.String("guild", e.GuildID), slog.String("channel", e.ChannelID))
		return
	} else if err != nil {
		h.log.Error("Error while trying to get channel from database",
			slog.String("guild", e.GuildID),
			slog.String("channel", e.ChannelID),
			slog.String("err", err.Error()),
		)
		return
	}

	gc, err := h.db.ChannelGroup(ch.GuildID, ch.ID)
	if errors.Is(err, guilddb.ErrNotFound) {
		h.log.Debug("Channel is not in a group, ignoring.", slog.String("guild", e.GuildID), slog.String("channel", e.ChannelID))
		return
	} else if err != nil {
		h.log.Error("Error while trying to get channel group from database",
			slog.String("guild", e.GuildID),
			slog.String("channel", e.ChannelID),
			slog.String("err", err.Error()),
		)
		return
	}

	_, err = h.getMessage(e.Message, ch.Language)
	if err != nil {
		h.log.Error("Error while trying to get/set message to database",
			slog.String("guild", e.Message.GuildID),
			slog.String("channel", e.Message.ChannelID),
			slog.String("message", e.Message.ID),
			slog.String("err", err.Error()),
		)
		_, err := s.ChannelMessageSendReply(
			e.Message.ChannelID,
			fmt.Sprintf("Error while trying to send message to database. %s", err.Error()),
			e.Message.Reference(),
		)
		if err != nil {
			h.log.Error("Error while trying to send error message",
				slog.String("guild", e.Message.GuildID),
				slog.String("channel", e.Message.ChannelID),
				slog.String("message", e.Message.ID),
				slog.String("err", err.Error()),
			)
		}
		return
	}

	for _, c := range gc {
		if c.ID == ch.ID && c.GuildID == ch.GuildID {
			continue
		}
		go func(c guilddb.Channel) {
			uw, err := h.getUserWebhook(s, c.ID, e.Message.Author)
			if err != nil {
				h.log.Error("Error while trying to create user webhook",
					slog.String("guild", e.Message.GuildID),
					slog.String("channel", e.Message.ChannelID),
					slog.Any("user", e.Message.Author),
				)
				_, err := s.ChannelMessageSendReply(
					e.Message.ChannelID,
					fmt.Sprintf("Error while trying to create user webhook %s", err.Error()),
					e.Message.Reference(),
				)
				if err != nil {
					h.log.Error("Error while trying to send error message",
						slog.String("guild", e.Message.GuildID),
						slog.String("channel", e.Message.ChannelID),
						slog.String("message", e.Message.ID),
						slog.String("err", err.Error()),
					)
				}
			}

			t, err := h.translator.Translate(ch.Language, c.Language, e.Message.Content)
			if err != nil {
				h.log.Error("Error while trying to translate message",
					slog.String("guild", e.Message.GuildID),
					slog.String("channel", e.Message.ChannelID),
					slog.String("message", e.Message.ID),
					slog.String("content", e.Message.Content),
					slog.String("err", err.Error()),
				)
				_, err := s.ChannelMessageSendReply(
					e.Message.ChannelID,
					fmt.Sprintf("Error while trying to translate message. %s", err.Error()),
					e.Message.Reference(),
				)
				if err != nil {
					h.log.Error("Error while trying to send error message",
						slog.String("guild", e.Message.GuildID),
						slog.String("channel", e.Message.ChannelID),
						slog.String("message", e.Message.ID),
						slog.String("err", err.Error()),
					)
				}
			}

			tdm, err := s.WebhookExecute(uw.ID, uw.Token, true, &dgo.WebhookParams{
				AvatarURL: e.Message.Author.AvatarURL(""),
				Username:  e.Message.Author.GlobalName,
				Content:   t,
			})
			// tdm, err := s.ChannelMessageSend(c.ID, t)
			if err != nil {
				h.log.Error("Error while trying to send translated message",
					slog.String("guild", e.Message.GuildID),
					slog.String("channel", e.Message.ChannelID),
					slog.String("message", e.Message.ID),
					slog.String("content", e.Message.Content),
					slog.String("err", err.Error()),
				)
				_, err := s.ChannelMessageSendReply(
					e.Message.ChannelID,
					fmt.Sprintf("Error while trying to send translated message. %s", err.Error()),
					e.Message.Reference(),
				)
				if err != nil {
					h.log.Error("Error while trying to send error message",
						slog.String("guild", e.Message.GuildID),
						slog.String("channel", e.Message.ChannelID),
						slog.String("message", e.Message.ID),
						slog.String("err", err.Error()),
					)
				}
			}

			if tdm.GuildID == "" {
				tdm.GuildID = e.Message.GuildID
			}

			_, err = h.getTranslatedMessage(tdm, e.Message, c.Language)
			if err != nil {
				h.log.Error("Error while trying to get/set translated message to database",
					slog.String("guild", e.Message.GuildID),
					slog.String("channel", e.Message.ChannelID),
					slog.String("message", e.Message.ID),
					slog.String("err", err.Error()),
				)
				_, err := s.ChannelMessageSendReply(
					e.Message.ChannelID,
					fmt.Sprintf("Error while trying to send translated message to database. %s", err.Error()),
					e.Message.Reference(),
				)
				if err != nil {
					h.log.Error("Error while trying to send error message",
						slog.String("guild", e.Message.GuildID),
						slog.String("channel", e.Message.ChannelID),
						slog.String("message", e.Message.ID),
						slog.String("err", err.Error()),
					)
				}
			}
		}(c)

	}

}

func (h MessageCreate) getUserWebhook(s *dgo.Session, channelID string, user *dgo.User) (*dgo.Webhook, error) {
	var whName = "DISLATE_USER_WEBHOOK_" + user.ID

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

func (h MessageCreate) getMessage(m *dgo.Message, lang lang.Language) (guilddb.Message, error) {
	msg, err := h.db.Message(m.GuildID, m.ChannelID, m.ID)

	if errors.Is(err, guilddb.ErrNotFound) {
		if err := h.db.MessageInsert(guilddb.NewMessage(m.GuildID, m.ChannelID, m.ID, lang)); err != nil {
			return guilddb.Message{}, err
		}
		msg, err = h.db.Message(m.GuildID, m.ChannelID, m.ID)
		if err != nil {
			return guilddb.Message{}, err
		}
	}

	return msg, nil

}

func (h MessageCreate) getTranslatedMessage(m, original *dgo.Message, lang lang.Language) (guilddb.Message, error) {
	msg, err := h.db.Message(m.GuildID, m.ChannelID, m.ID)

	if errors.Is(err, guilddb.ErrNotFound) {
		if err := h.db.MessageInsert(guilddb.NewTranslatedMessage(
			m.GuildID,
			m.ChannelID,
			m.ID,
			lang,
			original.ChannelID,
			original.ID,
		)); err != nil {
			return guilddb.Message{}, err
		}
		msg, err = h.db.Message(m.GuildID, m.ChannelID, m.ID)
		if err != nil {
			return guilddb.Message{}, err
		}
	} else if err != nil {
		return guilddb.Message{}, err
	}

	return msg, nil

}
