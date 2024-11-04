package events

import (
	e "errors"
	"log/slog"
	"slices"
	"sync"

	"forge.capytal.company/capytal/dislate/bot/events/errors"
	"forge.capytal.company/capytal/dislate/bot/gconf"
	"forge.capytal.company/capytal/dislate/translator"

	gdb "dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type EThreadCreate struct {
	db         gconf.DB
	translator translator.Translator
}

func NewEThreadCreate(db gconf.DB, t translator.Translator) EThreadCreate {
	return EThreadCreate{db, t}
}

func (h EThreadCreate) Serve(s *dgo.Session, ev *dgo.ThreadCreate) errors.EventErr {
	log := gconf.GetLogger(ev.GuildID, s, h.db)
	everr := errors.NewThreadCreateErr(s, ev, log)

	parentCh, err := h.db.Channel(ev.GuildID, ev.ParentID)
	if e.Is(err, gdb.ErrNotFound) {
		log.Debug("Parent channel of thread not in database, ignoring",
			slog.String("thread", ev.ID),
			slog.String("parent", ev.ParentID),
		)
		return nil
	}

	// INFO: Threads have the same ID as the origin message of them
	threadMsg, err := h.db.Message(ev.GuildID, ev.ParentID, ev.ID)

	var startMsg *dgo.Message

	// If no thread message is found in database, it is probably a thread started without
	// a source message or a forum post.
	if e.Is(err, gdb.ErrNotFound) {
		ms, err := s.ChannelMessages(ev.ID, 10, "", "", "")
		if err != nil {
			return everr.Join(e.New("Failed to get messages of thread"), err)
		} else if len(ms) == 0 {
			log.Debug("Failed to get messages of thread, empty slice returned, probably created by bot, ignoring",
				slog.String("thread", ev.ID),
				slog.String("parent", ev.ParentID),
			)
			return nil
		}

		threadMsg = gdb.NewMessage(ev.GuildID, ev.ParentID, ev.ID, parentCh.Language)
		startMsg = ms[0]

	} else if err != nil {
		return everr.Join(e.New("Failed to get thread starter message from database"), err)
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

	dth, err := s.Channel(ev.ID)
	if err != nil {
		return everr.Join(e.New("Failed to get discord thread"), err)
	} else if !dth.IsThread() {
		return everr.Join(e.New("Channel is not a thread"))
	}

	th := gdb.NewChannel(dth.GuildID, dth.ID, threadMsg.Language)
	if err := h.db.ChannelInsert(th); e.Is(err, gdb.ErrNoAffect) {
		if err = h.db.MessageInsert(threadMsg); err != nil && !e.Is(err, gdb.ErrNoAffect) {
			return everr.Join(e.New("Failed to add thread started message to database"), err)
		}
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to add thread channel to database"), err)
	}

	parentChannelGroup, err := h.db.ChannelGroup(parentCh.GuildID, parentCh.ID)
	if e.Is(err, gdb.ErrNotFound) {
		parentChannelGroup = gdb.ChannelGroup{parentCh}
	} else if err != nil {
		return everr.Join(e.New("Failed to get parent channel group"))
	}

	var wg sync.WaitGroup

	tg := make(chan gdb.Channel, len(parentChannelGroup))
	errs := make(chan errors.EventErr)

	for _, pc := range parentChannelGroup {
		if pc.ID == dth.ParentID {
			continue
		}

		m, err := h.db.MessageWithOriginByLang(pc.GuildID, pc.ID, originMsg.ID, pc.Language)
		if e.Is(err, gdb.ErrNotFound) && startMsg != nil {

			wg.Add(1)
			go func(pc gdb.Channel, tg chan<- gdb.Channel, errs chan<- errors.EventErr) {
				defer wg.Done()

				everr := errors.NewThreadCreateErr(s, ev, log)
				everr.AddData("TranslatedParentID", pc.ID)

				parentDCh, err := s.Channel(pc.ID)
				if err != nil {
					errs <- everr.Join(e.New("Failed to get translated parent channel object"), err)
					return
				}

				content, err := h.translator.Translate(
					parentCh.Language,
					pc.Language,
					startMsg.Content,
				)
				if err != nil {
					errs <- everr.Join(e.New("Failed to translate forum post of thread"), err)
					return
				}

				var dtth *dgo.Channel
				var msg *dgo.Message

				if parentDCh.Type == dgo.ChannelTypeGuildForum && startMsg != nil {
					tags := slices.DeleteFunc(dth.AppliedTags, func(t string) bool {
						return !slices.ContainsFunc(
							parentDCh.AvailableTags,
							func(pt dgo.ForumTag) bool {
								return pt.Name == t
							},
						)
					})

					dtth, err = s.ForumThreadStartComplex(pc.ID, &dgo.ThreadStart{
						Name:                dth.Name,
						AutoArchiveDuration: dth.ThreadMetadata.AutoArchiveDuration,
						Type:                dth.Type,
						Invitable:           dth.ThreadMetadata.Invitable,
						RateLimitPerUser:    dth.RateLimitPerUser,
						AppliedTags:         tags,
					}, &dgo.MessageSend{
						Content:    content,
						Embeds:     startMsg.Embeds,
						TTS:        startMsg.TTS,
						Components: startMsg.Components,
					})
					if err != nil {
						errs <- everr.Join(e.New("Failed to translate forum post of thread"), err)
						return
					}

					msg, err = s.ChannelMessage(dtth.ID, dtth.ID)
					if err != nil {
						errs <- everr.Join(e.New("Failed to get translated thread starter message"), err)
						return
					}

				} else {
					dtth, err = s.ThreadStartComplex(pc.ID, &dgo.ThreadStart{
						Name:                dth.Name,
						AutoArchiveDuration: dth.ThreadMetadata.AutoArchiveDuration,
						Type:                dth.Type,
						Invitable:           dth.ThreadMetadata.Invitable,
						RateLimitPerUser:    dth.RateLimitPerUser,
					})
					if err != nil {
						errs <- everr.Join(e.New("Failed to create thread"), err)
						return
					}

					uw, err := getUserWebhook(s, pc.ID, startMsg.Author)
					if err != nil {
						errs <- everr.Join(e.New("Failed to get/set user webhook for parent channel of translated thread"), err)
						return
					}

					msg, err = s.WebhookThreadExecute(uw.ID, uw.Token, true, dtth.ID, &dgo.WebhookParams{
						AvatarURL: startMsg.Author.AvatarURL(""),
						Username:  startMsg.Author.GlobalName,
						Content:   content,
					})
					if err != nil {
						errs <- everr.Join(e.New("Error while trying to execute user webhook"), err)
						return
					}
				}

				if err := h.db.ChannelInsert(gdb.NewChannel(dtth.GuildID, dtth.ID, pc.Language)); err != nil &&
					!e.Is(err, gdb.ErrNoAffect) {
					everr.AddData("TranslatedThreadID", dtth.ID)
					errs <- everr.Join(e.New("Failed to add translated thread to database"), err)
					return
				}

				err = h.db.MessageInsert(
					gdb.NewTranslatedMessage(
						dtth.GuildID,
						dtth.ID,
						msg.ID,
						pc.Language,
						startMsg.ChannelID,
						startMsg.ID,
					),
				)
				if err != nil {
					errs <- everr.Join(e.New("Failed to add translated thread starter message to database"), err)
					return
				}

				tg <- gdb.NewChannel(dtth.GuildID, dtth.ID, pc.Language)
			}(
				pc,
				tg,
				errs,
			)

		} else if err != nil {
			return everr.Join(e.New("Failed to get thread translated start message"), err)
		} else {

			wg.Add(1)
			go func(m gdb.Message, tg chan<- gdb.Channel, errs chan<- errors.EventErr) {
				defer wg.Done()
				everr := errors.NewThreadCreateErr(s, ev, log)
				everr.AddData("TranslatedParentID", m.ChannelID)

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
					errs <- everr.Join(e.New("Failed to create translated thread"), err)
					return
				}
				everr.AddData("TranslatedThreadID", pc.ID)

				if err := h.db.ChannelInsert(gdb.NewChannel(dtth.GuildID, dtth.ID, m.Language)); err != nil &&
					!e.Is(err, gdb.ErrNoAffect) {
					errs <- everr.Join(e.New("Failed to add translated thread to database"), err)
					return
				}

				tg <- gdb.NewChannel(dtth.GuildID, dtth.ID, m.Language)
			}(m, tg, errs)

		}
	}

	wg.Wait()

	everrs := make([]error, 0, len(errs))
	for err := range errs {
		everrs = append(everrs, err)
	}
	if len(errs) > 0 {
		return everr.Join(everrs...)
	}

	var threadGroup gdb.ChannelGroup
	for t := range tg {
		threadGroup = append(threadGroup, t)
	}

	if err := h.db.ChannelGroupInsert(threadGroup); err != nil {
		return everr.Join(e.New("Failed to add group of threads to database"), err)
	}

	thMsgs, err := s.ChannelMessages(th.ID, 10, "", "", "")
	if err != nil {
		return everr.Join(e.New("Failed to get thread messages"), err)
	}

	for _, m := range thMsgs {
		if startMsg != nil && m.ID == startMsg.ID {
			continue
		}
		if m.Content != "" {
			m.GuildID = th.GuildID
			NewMessageCreate(h.db, h.translator).sendMessage(log, s, m)
		}
	}

	return nil
}

type ThreadCreate struct {
	db         gconf.DB
	translator translator.Translator
	session    *dgo.Session
	thread     *dgo.Channel
	originLang translator.Language
}

func NewThreadCreate(db gconf.DB, t translator.Translator) ThreadCreate {
	return ThreadCreate{db, t, nil, nil, translator.EN}
}

func (h ThreadCreate) Serve(s *dgo.Session, ev *dgo.ThreadCreate) errors.EventErr {
	log := gconf.GetLogger(ev.GuildID, s, h.db)
	everr := errors.NewThreadCreateErr(s, ev, log)

	parentCh, err := h.db.Channel(ev.GuildID, ev.ParentID)
	if e.Is(err, gdb.ErrNotFound) {
		log.Debug("Parent channel of thread not in database, ignoring",
			slog.String("ThreadID", ev.ID),
			slog.String("ParentID", ev.ParentID))
		return nil
	}

	ms, err := s.ChannelMessages(ev.ID, 10, "", "", "")
	if err != nil {
		return everr.Join(e.New("Failed to get messages of thread"), err)
	} else if len(ms) == 0 || (len(ms) == 1 && ms[0].Type == dgo.MessageTypeThreadStarterMessage) {
		log.Debug("No messages found in thread, probably created by bot, ignoring",
			slog.String("ThreadID", ev.ID),
			slog.String("ParentID", ev.ParentID))
		return nil
	}

	// INFO: Threads have the same ID as their starter messages
	starterMsg, err := h.db.Message(parentCh.GuildID, parentCh.ID, ev.ID)
	if e.Is(err, gdb.ErrNotFound) {
		starterMsg = gdb.NewMessage(parentCh.GuildID, ev.ID, ev.ID, parentCh.Language)
		err = h.db.MessageInsert(starterMsg)
		if err != nil {
			return everr.Join(e.New("Failed to add starter message to database"), err)
		}
	}

	thread, err := s.Channel(starterMsg.ID)
	if err != nil {
		return everr.Join(e.New("Failed to get thread from discord"), err)
	} else if !thread.IsThread() {
		return everr.Join(e.New("Failed to get thread from discord, thread is not a thread somehow"), err)
	}

	parentChannelGroup, err := h.db.ChannelGroup(parentCh.GuildID, parentCh.ID)
	if e.Is(err, gdb.ErrNotFound) {
		log.Debug("Parent channel not in a group, ignoring",
			slog.String("ThreadID", ev.ID),
			slog.String("ParentID", ev.ParentID))
		return nil
	} else if err != nil {
		return everr.Join(e.New("Failed to get parent channel group"))
	}

	var wg sync.WaitGroup
	tg := make(chan gdb.Channel)
	errs := make(chan error)

	h.session = s
	h.originLang = parentCh.Language
	h.thread = thread

	for _, pc := range parentChannelGroup {
		if pc.ID == ev.ParentID {
			continue
		}

		wg.Add(1)
		go func(tg chan<- gdb.Channel, errs chan<- error) {
			defer wg.Done()
			t, err := h.startTranslatedThread(pc, starterMsg)
			tg <- t
			if err != nil {
				errs <- err
			}
			log.Debug("FINISHED")
		}(tg, errs)
	}

	wg.Wait()

	everrs := make([]error, 0, len(errs))
	for err := range errs {
		log.Debug("ERR")
		everrs = append(everrs, err)
	}
	if len(errs) > 0 {
		log.Debug("ERR RETURN")
		return everr.Join(everrs...)
	}

	log.Debug("FUNCTION 1")

	if err := h.db.ChannelInsert(gdb.NewChannel(thread.GuildID, thread.ID, parentCh.Language)); err != nil {
		return everr.Join(e.New("Failed to add thread channel to database"), err)
	}

	threadGroup := make(gdb.ChannelGroup, 0, len(tg))
	for t := range tg {
		threadGroup = append(threadGroup, t)
	}

	log.Debug("FUNCTION 2")

	if err := h.db.ChannelGroupInsert(threadGroup); err != nil {
		return everr.Join(e.New("Failed to add group of thread to database"), err)
	}

	thMsgs, err := s.ChannelMessages(thread.ID, 10, "", "", "")
	if err != nil {
		return everr.Join(e.New("Failed to get thread messages"), err)
	}

	log.Debug("FUNCTION 3")

	for _, m := range thMsgs {
		m.GuildID = thread.GuildID
		err := NewMessageCreate(h.db, h.translator).sendMessage(log, s, m)
		if err != nil {
			return everr.Join(e.New("Failed to translate thread messages"), err)
		}
	}

	return nil
}

func (h ThreadCreate) startTranslatedThread(
	pc gdb.Channel,
	sm gdb.Message,
) (gdb.Channel, error) {
	if sm.OriginChannelID != nil && *sm.OriginChannelID == pc.ID {
		m, err := h.db.Message(sm.GuildID, *sm.OriginChannelID, *sm.OriginID)
		if err != nil {
			return gdb.Channel{}, e.Join(
				e.New("Failed to get origin message of starter message"),
				err,
			)
		}
		return h.startTranslatedMessageThread(m)
	}

	m, err := h.db.MessageWithOriginByLang(sm.GuildID, sm.ChannelID, sm.ID, pc.Language)
	if e.Is(err, gdb.ErrNotFound) {
	} else if err != nil {
		return gdb.Channel{}, e.Join(
			e.New("Failed to get translated message of starter message"),
			err,
		)
	}

	return h.startTranslatedMessageThread(m)
}

func (h ThreadCreate) startTranslatedMessageThread(
	m gdb.Message,
) (gdb.Channel, error) {
	name, err := h.translator.Translate(h.originLang, m.Language, h.thread.Name)
	if err != nil {
		return gdb.Channel{}, e.Join(e.New("Failed to translate thread name"), err)
	}

	th, err := h.session.MessageThreadStartComplex(m.ChannelID, m.ID, &dgo.ThreadStart{
		Name:                name,
		AutoArchiveDuration: h.thread.ThreadMetadata.AutoArchiveDuration,
		Type:                h.thread.Type,
		Invitable:           h.thread.ThreadMetadata.Invitable,
		RateLimitPerUser:    h.thread.RateLimitPerUser,
		AppliedTags:         h.thread.AppliedTags,
	})
	if err != nil {
		return gdb.Channel{}, e.Join(e.New("Failed to create thread"), err)
	}

	c := gdb.NewChannel(th.GuildID, th.ID, m.Language)
	if err := h.db.ChannelInsert(c); err != nil {
		return c, e.Join(e.New("Failed to insert thread on database"), err)
	}

	return c, nil
}
