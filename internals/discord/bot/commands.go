package bot

import (
	"dislate/internals/discord/bot/commands"
	"fmt"
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
)

func (b *Bot) registerCommands() error {
	cs := []commands.Command{
		commands.NewManageChannel(b.db),
	}

	rcs := make([]*dgo.ApplicationCommand, len(cs))
	handlers := make(map[string]func(*dgo.Session, *dgo.InteractionCreate), len(cs))

	for i, v := range cs {
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v.Info())
		if err != nil {
			return err
		}

		handlers[cmd.Name] = func(s *dgo.Session, ic *dgo.InteractionCreate) {
			err := v.Handle(s, ic)
			if err != nil {
				_ = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
					Type: dgo.InteractionResponseDeferredChannelMessageWithSource,
					Data: &dgo.InteractionResponseData{
						Content: fmt.Sprintf("Error while trying to handle command: %s", err.Error()),
						Flags:   dgo.MessageFlagsEphemeral,
					},
				})
				b.logger.Error("Failed to handle command",
					slog.String("name", cmd.Name),
					slog.String("id", cmd.ID),
					slog.String("err", err.Error()),
				)
			}
		}
		rcs[i] = cmd

		b.logger.Info("Registered command",
			slog.String("name", cmd.Name),
			slog.String("id", cmd.ID),
		)
	}

	b.session.AddHandler(func(s *dgo.Session, i *dgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	return nil
}

func (b *Bot) removeCommands() error {
	cmds, err := b.session.ApplicationCommands(b.session.State.Application.ID, "")
	if err != nil {
		return err
	}

	for _, v := range cmds {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			return err
		}
		b.logger.Info("Removed command",
			slog.String("name", v.Name),
			slog.String("id", v.ID),
		)
	}
	return nil
}
