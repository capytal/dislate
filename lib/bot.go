package bot

import (
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
}

func Start(token string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	if err := s.Open(); err != nil {
		return nil, err
	}

	bot := &Bot{
		session:                s,
	}

	return bot, nil
}

func (b *Bot) Stop() error {
	if err := b.session.Close(); err != nil {
		return err
	}
	return nil
}

func (b *Bot) HandleCommand(c Command) {
	b.commands = append(b.commands, c)
}
