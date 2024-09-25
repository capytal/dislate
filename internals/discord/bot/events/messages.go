package events

import (
	"dislate/internals/discord/bot/events/errors"
	"dislate/internals/discord/bot/gconf"
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"dislate/internals/translator/lang"
	e "errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	dgo "github.com/bwmarrin/discordgo"
)

type MessageCreate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewMessageCreate(db gconf.DB, t translator.Translator) MessageCreate {
	return MessageCreate{db, t}
}

func (h MessageCreate) Serve(
	s *dgo.Session,
	ev *dgo.MessageCreate,
) errors.EventErr {
	if ev.Message.Author.Bot || (ev.Type != dgo.MessageTypeDefault && ev.Type != dgo.MessageTypeReply) {
		return nil
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)
	return h.sendMessage(log, s, ev.Message)
}

func (h MessageCreate) sendMessage(
	log *slog.Logger,
	s *dgo.Session,
	msg *dgo.Message,
) errors.EventErr {
	everr := errors.NewMessageErr[*dgo.MessageCreate](s, msg, log)

	ch, err := h.db.Channel(msg.GuildID, msg.ChannelID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in database, ignoring.",
			slog.String("guild", msg.GuildID),
			slog.String("channel", msg.ChannelID),
			slog.String("message", msg.ID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get channel from database"), err)
	}

	gc, err := h.db.ChannelGroup(ch.GuildID, ch.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Channel is not in a group, ignoring.",
			slog.String("guild", msg.GuildID),
			slog.String("channel", msg.ChannelID),
			slog.String("message", msg.ID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get channel group from database"), err)
	}

	_, err = getMessage(h.db, msg, ch.Language)
	if err != nil {
		return everr.Join(e.New("Failed to get/add message to database"), err)
	}

	var wg sync.WaitGroup
	errs := make(chan errors.EventErr)

	for _, c := range gc {
		if c.ID == ch.ID && c.GuildID == ch.GuildID {
			continue
		}
		wg.Add(1)
		go func(c guilddb.Channel, errs chan<- errors.EventErr) {
			defer wg.Done()

			everr := errors.NewMessageErr[*dgo.MessageCreate](s, msg, log)
			everr.AddData("TranslatedChannelID", c.ID)

			dch, err := s.Channel(c.ID)

			var channelID string
			if err != nil {
				errs <- everr.Join(e.New("Failed to get information about translated channel"), err)
				return
			} else if dch.IsThread() {
				channelID = dch.ParentID
			} else {
				channelID = dch.ID
			}

			uw, err := getUserWebhook(s, channelID, msg.Author)
			if err != nil {
				errs <- everr.Join(e.New("Failed to get/set user webhook for translated channel"), err)
				return
			}

			t, err := h.translator.Translate(ch.Language, c.Language, msg.Content)
			if err != nil {
				errs <- everr.Join(e.New("Error while trying to translate message"), err)
				return
			}

			if msg.Type == dgo.MessageTypeReply {
				t = createReply(msg, t)
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
				everr.AddData("WebhookID", uw.ID)
				errs <- everr.Join(e.New("Error while trying to execute user webhook"), err)
				return
			}

			if tdm.GuildID == "" {
				tdm.GuildID = msg.GuildID
			}

			_, err = getTranslatedMessage(h.db, tdm, msg, c.Language)
			if err != nil {
				everr.AddData("WebhookID", uw.ID)
				everr.AddData("TranslatedMessageID", uw.ID)
				errs <- everr.Join(e.New("Error while trying to add translated message to dabase"), err)
				return
			}
		}(c, errs)

	}

	wg.Wait()
	for err := range errs {
		everr.Join(err)
	}
	if len(errs) > 0 {
		return everr
	}

	return nil
}

func getMessageLink(msg *dgo.MessageReference) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", msg.GuildID, msg.ChannelID, msg.MessageID)
}

func createReply(msg *dgo.Message, t string) string {
	msgThreshold := 100
	if len(msg.ReferencedMessage.Content) < 100 {
		msgThreshold = len(msg.ReferencedMessage.Content)
	}
	// ↩️ or ➡️ ??
	replyMessage := fmt.Sprintf("↩️<@%s>: [`%s...`](%s)\n%s",
		msg.ReferencedMessage.Author.ID,
		msg.ReferencedMessage.Content[:msgThreshold],
		getMessageLink(msg.MessageReference),
		t)

	return replyMessage
}

type MessageUpdate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewMessageUpdate(db gconf.DB, t translator.Translator) MessageUpdate {
	return MessageUpdate{db, t}
}

func (h MessageUpdate) Serve(s *dgo.Session, ev *dgo.MessageUpdate) errors.EventErr {
	if ev.Message.Author.Bot || (ev.Type != dgo.MessageTypeDefault && ev.Type != dgo.MessageTypeReply) {
		return nil
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)
	everr := errors.NewMessageErr[*dgo.MessageUpdate](s, ev.Message, log)

	msg, err := h.db.Message(ev.Message.GuildID, ev.Message.ChannelID, ev.Message.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Message is not in database, ignoring.",
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get message from database"), err)
	}

	tmsgs, err := h.db.MessagesWithOrigin(msg.GuildID, msg.ChannelID, msg.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("No translated message found, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get translated messages from database"), err)
	}

	var wg sync.WaitGroup
	errs := make(chan errors.EventErr)

	for _, m := range tmsgs {
		if m.ID == msg.ID && m.GuildID == msg.GuildID {
			continue
		}
		wg.Add(1)
		go func(m guilddb.Message, errs chan<- errors.EventErr) {
			defer wg.Done()

			everr := errors.NewMessageErr[*dgo.MessageUpdate](s, ev.Message, log)
			everr.AddData("TranslatedMessageID", m.ID)
			everr.AddData("TranslatedChannelID", m.ChannelID)

			var channelID string
			if dch, err := s.Channel(m.ChannelID); err != nil {
				errs <- everr.Join(e.New("Failed to get information about translated channel"), err)
				return
			} else if dch.IsThread() {
				channelID = dch.ParentID
			} else {
				channelID = dch.ID
			}

			uw, err := getUserWebhook(s, channelID, ev.Message.Author)
			if err != nil {
				errs <- everr.Join(e.New("Failed to get/set user webhook for translated channel"), err)
				return
			}

			t, err := h.translator.Translate(msg.Language, m.Language, ev.Message.Content)
			if err != nil {
				errs <- everr.Join(e.New("Error while trying to translate message"), err)
				return
			}

			_, err = s.WebhookMessageEdit(uw.ID, uw.Token, m.ID, &dgo.WebhookEdit{
				Content: &t,
			})
			if err != nil {
				everr.AddData("WebhookID", uw.ID)
				errs <- everr.Join(e.New("Error while trying to execute user webhook"), err)
				return
			}
		}(m, errs)

	}

	wg.Wait()
	for err := range errs {
		everr.Join(err)
	}
	if len(errs) > 0 {
		return everr
	}

	return nil
}

type MessageDelete struct {
	db gconf.DB
}

func NewMessageDelete(db gconf.DB) MessageDelete {
	return MessageDelete{db}
}

func (h MessageDelete) Serve(s *dgo.Session, ev *dgo.MessageDelete) errors.EventErr {
		if ev.Type != dgo.MessageTypeDefault && ev.Type != dgo.MessageTypeReply {
		return nil
	}

	log := gconf.GetLogger(ev.Message.GuildID, s, h.db)
	everr := errors.NewMessageErr[*dgo.MessageUpdate](s, ev.Message, log)

	msg, err := h.db.Message(ev.Message.GuildID, ev.Message.ChannelID, ev.Message.ID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("Message is not in database, ignoring.",
			slog.String("guild", ev.Message.GuildID),
			slog.String("channel", ev.Message.ChannelID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get message from database"), err)
	}

	var originChannelID, originID string
	if msg.OriginID != nil && msg.OriginChannelID != nil {
		oMsg, err := h.db.Message(ev.Message.GuildID, *msg.OriginChannelID, *msg.OriginID)
		if err != nil {
			originChannelID = *msg.OriginChannelID
			originID = *msg.OriginID
		} else {
			msg = oMsg
			originChannelID = oMsg.ChannelID
			originID = oMsg.ID
		}
	} else {
		originChannelID = msg.ChannelID
		originID = msg.ID
	}

	tmsgs, err := h.db.MessagesWithOrigin(msg.GuildID, originChannelID, originID)
	if e.Is(err, guilddb.ErrNotFound) {
		log.Debug("No translated message found, ignoring.",
			slog.String("guild", ev.GuildID),
			slog.String("channel", ev.ChannelID),
		)
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get translated messages from database"), err)
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

	var wg sync.WaitGroup
	errs := make(chan errors.EventErr)

	for _, m := range append(tmsgs, msg) {
		go func(m guilddb.Message, errs chan<- errors.EventErr) {
			everr := errors.NewMessageErr[*dgo.MessageUpdate](s, ev.Message, log)
			everr.AddData("TranslatedMessageID", m.ID)
			everr.AddData("TranslatedChannelID", m.ChannelID)

			err := h.db.MessageDeleteFromChannel(guilddb.NewChannel(m.GuildID, m.ID, lang.EN))
			if err != nil && !e.Is(err, guilddb.ErrNoAffect) {
				errs <- everr.Join(e.New("Failed to delete message from channel"), err)
				return
			}

			err = h.db.ChannelDelete(guilddb.NewChannel(m.GuildID, m.ID, lang.EN))
			if err != nil && !e.Is(err, guilddb.ErrNoAffect) {
				errs <- everr.Join(e.New("Failed to delete message thread from channel"), err)
				return
			}
		}(m, errs)
	}

	wg.Wait()
	for err := range errs {
		everr.Join(err)
	}
	if len(errs) > 0 {
		return everr
	}

	if err := h.db.MessageDelete(guilddb.NewMessage(msg.GuildID, msg.ChannelID, msg.ID, lang.EN)); err != nil {
		return everr.Join(e.New("Failed to delete message from database"), err)
	}

	return nil
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
