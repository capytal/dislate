package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"forge.capytal.company/capytal/dislate/bot/commands"

	dgo "github.com/bwmarrin/discordgo"
)

func (b *Bot) registerCommands() error {
	cs := []commands.Command{
		commands.NewMagageConfig(b.db),
		commands.NewManageChannel(b.db),
	}

	handlers := make(map[string]func(*dgo.Session, *dgo.InteractionCreate), len(cs))
	componentsHandlers := make(map[string]func(*dgo.Session, *dgo.InteractionCreate))

	for _, v := range cs {
		var cmd *dgo.ApplicationCommand
		var err error
		subCmds := make(map[string]commands.Command)

		sb := v.Subcommands()

		if len(sb) == 0 {
			cmd, err = b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v.Info())
			if err != nil {
				return err
			}
		} else {
			subCmdsOpts := make([]*dgo.ApplicationCommandOption, len(sb))
			for i, sb := range sb {
				subCmds[sb.Info().Name] = sb
				subCmdsOpts[i] = &dgo.ApplicationCommandOption{
					Type:        dgo.ApplicationCommandOptionSubCommand,
					Name:        sb.Info().Name,
					Description: sb.Info().Description,
					Options:     sb.Info().Options,
				}
			}
			info := v.Info()
			info.Options = subCmdsOpts

			cmd, err = b.session.ApplicationCommandCreate(b.session.State.User.ID, "", info)
			if err != nil {
				return err
			}
		}

		for _, c := range v.Components() {
			cj, err := c.Info().MarshalJSON()
			if err != nil {
				return errors.Join(fmt.Errorf("Failed to marshal command"), err)
			}

			var v struct {
				CustomID string `json:"custom_id"`
			}
			if err := json.Unmarshal(cj, &v); err != nil {
				return errors.Join(fmt.Errorf("Failed to unmarshal command"), err)
			}

			componentsHandlers[v.CustomID] = func(s *dgo.Session, ic *dgo.InteractionCreate) {
				b.logger.Debug("Handling message component",
					slog.String("id", ic.Interaction.ID),
					slog.String("custom_id", ic.Interaction.MessageComponentData().CustomID),
				)
				err := c.Handle(s, ic)
				if err != nil {
					b.logger.Error("Failed to handle message component",
						slog.String("custom_id", ic.Interaction.MessageComponentData().CustomID),
						slog.String("err", err.Error()),
					)
				}
			}
		}

		handlers[cmd.Name] = func(s *dgo.Session, ic *dgo.InteractionCreate) {
			b.logger.Debug("Handling command",
				slog.String("id", ic.Interaction.ID),
				slog.String("name", ic.Interaction.ApplicationCommandData().Name),
			)

			opts := ic.Interaction.ApplicationCommandData().Options
			isSub := slices.IndexFunc(
				opts,
				func(o *dgo.ApplicationCommandInteractionDataOption) bool {
					return o.Type == dgo.ApplicationCommandOptionSubCommand
				},
			)
			if isSub != -1 {
				sc := opts[isSub]

				err := subCmds[sc.Name].Handle(s, ic)
				if err != nil {
					_ = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
						Type: dgo.InteractionResponseDeferredChannelMessageWithSource,
						Data: &dgo.InteractionResponseData{
							Content: fmt.Sprintf(
								"Error while trying to handle sub command: %s",
								err.Error(),
							),
							Flags: dgo.MessageFlagsEphemeral,
						},
					})
					b.logger.Error("Failed to handle sub command",
						slog.String("name", sc.Name),
						slog.String("err", err.Error()),
					)
				}

				return
			}

			err := v.Handle(s, ic)
			if err != nil {
				_ = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
					Type: dgo.InteractionResponseDeferredChannelMessageWithSource,
					Data: &dgo.InteractionResponseData{
						Content: fmt.Sprintf(
							"Error while trying to handle command: %s",
							err.Error(),
						),
						Flags: dgo.MessageFlagsEphemeral,
					},
				})
				b.logger.Error("Failed to handle command",
					slog.String("name", cmd.Name),
					slog.String("id", cmd.ID),
					slog.String("err", err.Error()),
				)
			}
		}

		b.logger.Info("Registered command",
			slog.String("name", cmd.Name),
			slog.String("id", cmd.ID),
		)
	}

	b.session.AddHandler(func(s *dgo.Session, i *dgo.InteractionCreate) {
		switch i.Interaction.Type {
		case dgo.InteractionApplicationCommand:
			if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case dgo.InteractionMessageComponent:
			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
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
