package bot

import (
	"dislate/internals/discord/bot/commands"

	dgo "github.com/bwmarrin/discordgo"
)

func (b *Bot) registerCommands() error {
	cs := []commands.Command{
		commands.NewTest(b.translator),
	}

	rcs := make([]*dgo.ApplicationCommand, len(cs))
	handlers := make(map[string]func(*dgo.Session, *dgo.InteractionCreate), len(cs))

	for i, v := range cs {
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v.Info())
		if err != nil {
			return err
		}

		handlers[cmd.Name] = v.Handle
		rcs[i] = cmd
	}

	b.session.AddHandler(func(s *dgo.Session, i *dgo.InteractionCreate) {
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
