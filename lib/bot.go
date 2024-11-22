package bot

import (
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session  *discordgo.Session
}

func New(token string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	return &Bot{
		session:  s,
		commands: []Command{},
	}, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	return nil
}
