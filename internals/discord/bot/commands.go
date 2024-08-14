package bot

import (
	"dislate/internals/translator/lang"

	"github.com/bwmarrin/discordgo"
)

type command struct {
	command *discordgo.ApplicationCommand
	handler func(*discordgo.Session, *discordgo.InteractionCreate)
}

func (b *Bot) commands() []command {
	cmds := []command{{
		&discordgo.ApplicationCommand{
			Name:        "test-command",
			Description: "This is a test command",
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			txt, _ := b.translator.Translate(lang.EN, lang.PT, "Hello world!")

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: txt,
				},
			})
		},
	}}
	return cmds
}

func (b *Bot) registerCommands() error {
	cs := b.commands()
	rcs := make([]*discordgo.ApplicationCommand, len(cs))
	handlers := make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate), len(cs))

	for i, v := range cs {
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v.command)
		if err != nil {
			return err
		}

		handlers[cmd.Name] = v.handler
		rcs[i] = cmd
	}

	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	return nil
}

func (b *Bot) removeCommands() error {
	for _, v := range b.registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
