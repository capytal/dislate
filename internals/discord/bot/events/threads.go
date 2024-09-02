package events

import (
	"dislate/internals/discord/bot/errors"
	"dislate/internals/discord/bot/gconf"
	"dislate/internals/translator"
	e "errors"
	"log/slog"
	"sync"

	gdb "dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type ThreadCreate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewThreadCreate(db gconf.DB, t translator.Translator) ThreadCreate {
	return ThreadCreate{db, t}
}

func (h ThreadCreate) Serve(s *dgo.Session, ev *dgo.ThreadCreate) {
	log := gconf.GetLogger(ev.GuildID, s, h.db)
	log.Debug("Thread created!", slog.String("parent", ev.ParentID), slog.String("thread", ev.ID))

	evErr := errors.NewEventError[ThreadCreate](map[string]any{
		"ThreadID": ev.ID,
		"ParentID": ev.ParentID,
		"GuildID":  ev.GuildID,
	})

	// INFO: Threads have the same ID as the origin message of them
	threadMsg, err := h.db.Message(ev.GuildID, ev.ParentID, ev.ID)
	if e.Is(err, gdb.ErrNotFound) {
		log.Debug("Parent message of thread not in database, ignoring",
			slog.String("thread", ev.ID),
			slog.String("parent", ev.ParentID),
			slog.String("error", err.Error()),
		)
		return
	} else if err != nil {
		evErr.Wrap(e.New("Failed to get thread message"), err).Log(log).Send(s, ev.ID)
		return
	}

	var originMsg gdb.Message
	if threadMsg.OriginID != nil && threadMsg.OriginChannelID != nil {
		oMsg, err := h.db.Message(ev.GuildID, *threadMsg.OriginChannelID, *threadMsg.OriginID)
		if err != nil {
			originMsg = threadMsg
		} else {
			originMsg = oMsg
		}
	} else {
		originMsg = threadMsg
	}

	msgs, err := h.db.MessagesWithOrigin(ev.GuildID, originMsg.ChannelID, originMsg.ID)
	if e.Is(err, gdb.ErrNotFound) {
		log.Debug("No translated messages for thread parent message found, ignoring",
			slog.String("thread message", ev.ID),
			slog.String("parent channel", ev.ParentID),
		)
		return
	} else if err != nil {
		evErr.Wrapf("Failed to get parent's translated messagas", err).
			AddData("OriginMessageID", originMsg.ID).
			AddData("OriginChannelID", originMsg.ChannelID).
			Log(log).
			Send(s, ev.ID)
		return
	}
	msgs = append(msgs, originMsg)

	dth, err := s.Channel(ev.ID)
	if err != nil {
		evErr.Wrapf("Failed to get discord thread", err).Log(log).Send(s, ev.ID)
		return
	} else if !dth.IsThread() {
		evErr.Wrapf("Channel is not a thread").Log(log).Send(s, ev.ID)
		return
	}

	th := gdb.NewChannel(dth.GuildID, dth.ID, threadMsg.Language)
	if err := h.db.ChannelInsert(th); e.Is(err, gdb.ErrNoAffect) {
		log.Info("Thread already in database, probably created by bot",
			slog.String("thread", dth.ID),
			slog.String("parent", dth.ParentID),
		)
		return
	} else if err != nil {
		evErr.Wrapf("Failed to add thread channel to database", err).Log(log).Send(s, ev.ID)
		return
	}

	threadGroup := make([]gdb.Channel, len(msgs))

	var wg sync.WaitGroup

	for i, m := range msgs {

		threadGroup[i] = gdb.NewChannel(m.GuildID, m.ID, m.Language)

		if m.ID == th.ID {
			continue
		}

		wg.Add(1)

		go func(m gdb.Message) {
			defer wg.Done()

			dtth, err := s.MessageThreadStartComplex(
				m.ChannelID,
				m.ID,
				&dgo.ThreadStart{
					Name:                dth.Name,
					AutoArchiveDuration: dth.ThreadMetadata.AutoArchiveDuration,
					Type:                dth.Type,
					Invitable:           dth.ThreadMetadata.Invitable,
					RateLimitPerUser:    dth.RateLimitPerUser,
					AppliedTags:         dth.AppliedTags,
				},
			)
			if err != nil {
				evErr.Wrapf("Failed to create translated thread", err).Log(log).Send(s, ev.ID)
				return
			}

			if err := h.db.ChannelInsert(gdb.NewChannel(dtth.GuildID, dtth.ID, m.Language)); err != nil &&
				!e.Is(err, gdb.ErrNoAffect) {
				evErr.Wrapf("Failed to add translated thread to database", err).
					AddData("TranslatedThreadID", dtth.ID).
					AddData("TranslatedParentID", dtth.ParentID).
					Log(log).
					Send(s, ev.ID)
				return
			}
		}(m)
	}

	wg.Wait()

	if err := h.db.ChannelGroupInsert(threadGroup); err != nil {
		evErr.Wrapf("Failed to add group of threads to database", err).
			AddData("ThreadGroup", threadGroup).
			Log(log).
			Send(s, ev.ID)
		return
	}

	thMsgs, err := s.ChannelMessages(th.ID, 10, "", "", "")
	if err != nil {
		evErr.Wrapf("Failed to get thread messages", err).Log(log).Send(s, ev.ID)
		return
	}

	for _, m := range thMsgs {
		if m.Content != "" {
			m.GuildID = th.GuildID
			NewMessageCreate(h.db, h.translator).sendMessage(log, s, m)
		}
	}
}
