package discord

import (
	"dislate/internals/guilddb"
	"dislate/internals/translator"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	token      string
	db         guilddb.GuildDB
	translator translator.Translator
	session    *discordgo.Session
}

func NewBot(token string, db guilddb.GuildDB, translator translator.Translator) (*Bot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return &Bot{}, err
	}

	return &Bot{token, db, translator, discord}, nil
}

func (b *Bot) Start() error {
	return b.session.Open()
}
func (b *Bot) Stop() error {
	return b.session.Open()
}
