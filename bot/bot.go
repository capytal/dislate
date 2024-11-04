package bot

import (
	"database/sql"
	"log/slog"

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
	s, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}

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
	return b.session.Open()
}

func (b *Bot) Stop() error {
	return b.session.Close()
}
