package bot

import (
	"log/slog"

	"forge.capytal.company/capytal/dislate/translator"

	"forge.capytal.company/capytal/dislate/bot/gconf"

	dgo "github.com/bwmarrin/discordgo"
)

type Bot struct {
	token      string
	db         gconf.DB
	translator translator.Translator
	session    *dgo.Session
	logger     *slog.Logger
}

func NewBot(
	token string,
	db gconf.DB,
	translator translator.Translator,
	logger *slog.Logger,
) (*Bot, error) {
	discord, err := dgo.New("Bot " + token)
	if err != nil {
		return &Bot{}, err
	}

	return &Bot{
		token:      token,
		db:         db,
		translator: translator,
		session:    discord,
		logger:     logger,
	}, nil
}

func (b *Bot) Start() error {
	b.registerEventHandlers()

	b.session.Identify.Intents = dgo.MakeIntent(dgo.IntentsAllWithoutPrivileged)

	if err := b.session.Open(); err != nil {
		return err
	}

	if err := b.registerCommands(); err != nil {
		return err
	}
	return nil
}

func (b *Bot) Stop() error {
	if err := b.removeCommands(); err != nil {
		return err
	}
	return b.session.Close()
}
