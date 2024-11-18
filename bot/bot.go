package bot

import (
	"database/sql"
	"errors"
	"log/slog"

	"forge.capytal.company/capytal/dislate/commands"
	"forge.capytal.company/capytal/dislate/db"
	"forge.capytal.company/capytal/dislate/translator"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	translator translator.Translator
	db         *db.Queries
	logger     *slog.Logger
	session    *discordgo.Session
}

func NewBot(
	token string,
	database *sql.DB,
	translator translator.Translator,
	log *slog.Logger,
) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	db, err := db.Prepare(database)
	if err != nil {
		return nil, err
	}

	return &Bot{
		session:    s,
		db:         db,
		translator: translator,
		logger:     log,
	}, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	ch := commands.NewCommandsHandler(b.logger, b.session)

	COMMANDS := []commands.Command{
		&mockCommand{},
	}

	if err := ch.UpdateCommands(COMMANDS); err != nil {
		return errors.Join(errors.New("Failed to update commands"), err)
	}

	return nil
}

func (b *Bot) Stop() error {
	return b.session.Close()
}
