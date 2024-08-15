package bot

import (
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	token              string
	db                 guilddb.GuildDB
	translator         translator.Translator
	session            *discordgo.Session
	logger             *slog.Logger
	registeredCommands []*discordgo.ApplicationCommand
}

func NewBot(token string, db guilddb.GuildDB, translator translator.Translator, logger *slog.Logger) (*Bot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return &Bot{}, err
	}

	return &Bot{
		token:              token,
		db:                 db,
		translator:         translator,
		session:            discord,
		logger:             logger,
		registeredCommands: make([]*discordgo.ApplicationCommand, 0),
	}, nil
}

func (b *Bot) Start() error {
	b.registerEventHandlers()

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
