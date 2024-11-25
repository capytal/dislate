package bot

import (
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session

	chatCommandHandlers    chatCommandHandlers
	messageCommandHandlers messageCommandHandlers
	userCommandHandlers    userCommandHandlers
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
		chatCommandHandlers:    chatCommandHandlers{},
		messageCommandHandlers: messageCommandHandlers{},
		userCommandHandlers:    userCommandHandlers{},
	}

	bot.session.AddHandler(bot.handleInteraction)

	return bot, nil
}

func (b *Bot) Stop() error {
	if err := b.session.Close(); err != nil {
		return err
	}
	return nil
}

func (b *Bot) handleInteraction(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	var err error

	switch ic.Type {
	case discordgo.InteractionApplicationCommand:
		err = b.handleApplicationCommand(s, ic)
	}

	_, err = s.ChannelMessageSend(ic.ChannelID, err.Error())
	panic(err) // TODO: handle error
}

func (b *Bot) handleApplicationCommand(
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
) error {
	data := ic.ApplicationCommandData()

	var err error

	if data.CommandType == discordgo.ChatApplicationCommand {
		err = b.chatCommandHandlers[sessionCommandID{data.ID, ic.GuildID}].Handle(
			ChatCommandCtx{},
		)
	} else if data.CommandType == discordgo.MessageApplicationCommand {
		err = b.messageCommandHandlers[sessionCommandID{data.ID, ic.GuildID}].Handle(MessageCommandCtx{})
	} else if data.CommandType == discordgo.UserApplicationCommand {
		err = b.userCommandHandlers[sessionCommandID{data.ID, ic.GuildID}].Handle(UserCommandCtx{})
	}

	return err
}
