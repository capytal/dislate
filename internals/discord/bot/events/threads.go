package events

import (
	"dislate/internals/discord/bot/errors"
	"dislate/internals/discord/bot/gconf"
	gdb "dislate/internals/guilddb"
	"dislate/internals/translator"
	e "errors"
	"log/slog"
	"sync"

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

	if len(ev.AppliedTags) > 0 {
		log.Debug("New thread is in forum, unimplemented, ignoring")
		return
	}

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
		errors.New("Unable to get thread parent message from database",
			slog.String("thread", ev.ID),
			slog.String("parent", ev.ParentID),
			slog.String("error", err.Error()),
		).Log(log)
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
		errors.NewErrDatabase(
			slog.String("thread message", ev.ID),
			slog.String("parent channel", ev.ParentID),
			slog.String("error", err.Error()),
		).LogSend(log, s, ev.ParentID)
		return
	}
	msgs = append(msgs, originMsg)

	dth, err := s.Channel(ev.ID)
	if err != nil {
		errors.New("Failed to get message thread object",
			slog.String("thread", ev.ID),
			slog.String("parent", ev.ParentID),
			slog.String("error", err.Error()),
		).LogSend(log, s, ev.ParentID)
		return
	} else if !dth.IsThread() {
		errors.New("Channel is not a thread for some reason",
			slog.String("channel", ev.ID),
			slog.String("parent", ev.ParentID),
		).LogSend(log, s, ev.ID)
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
		errors.New("Failed to add origin thread channel to database",
			slog.String("thread", dth.ID),
			slog.String("parent", dth.ParentID),
			slog.String("err", err.Error()),
		).LogSend(log, s, ev.ID)
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
				errors.New("Failed to create translated thread",
					slog.String("origin thread", dth.ID),
					slog.String("origin thread parent", dth.ParentID),
					slog.String("error", err.Error()),
				).LogSend(log, s, ev.ID)
				return
			}

			if err := h.db.ChannelInsert(gdb.NewChannel(dtth.GuildID, dtth.ID, m.Language)); err != nil &&
				!e.Is(err, gdb.ErrNoAffect) {
				errors.New("Failed to add thread channel to database",
					slog.String("thread", dth.ID),
					slog.String("parent", dth.ParentID),
					slog.String("origin thread", dth.ID),
					slog.String("origin thread parent", dth.ParentID),
					slog.String("err", err.Error()),
				).LogSend(log, s, dtth.ParentID)
				return
			}
		}(m)
	}

	wg.Wait()

	if err := h.db.ChannelGroupInsert(threadGroup); err != nil {
		errors.New("Failed to insert group of threads in database",
			slog.String("origin thread", dth.ID),
			slog.String("origin thread parent", dth.ParentID),
			slog.Any("thread group", threadGroup),
			slog.String("error", err.Error()),
		).LogSend(log, s, ev.ID)
		return
	}

	thMsgs, err := s.ChannelMessages(th.ID, 10, "", "", "")
	if err != nil {
		errors.New("Failed to get thread messages",
			slog.String("thread", dth.ID),
			slog.String("parent", dth.ParentID),
			slog.String("err", err.Error()),
		).LogSend(log, s, ev.ID)
		return
	}

	for _, m := range thMsgs {
		if m.Content != "" {
			m.GuildID = th.GuildID
			NewMessageCreate(h.db, h.translator).sendMessage(log, s, m)
		}
	}
}
