package commands

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func (h *CommandsHandler) handleInteraction(
	cmdsFunc map[CommandName]CommandFunc,
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
) {
	var userID string
	if ic.User != nil {
		userID = ic.User.ID
	} else {
		userID = ic.Member.User.ID
	}

	switch ic.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(cmdsFunc, s, ic)
	case discordgo.InteractionMessageComponent:
		// TODO!
	default:
		h.logger.Error("Application interaction is not a supported type!",
			slog.String("interaction_id", ic.ID),
			slog.String("interaction_type", ic.Type.String()),
			slog.String("interaction_user_id", userID),
			slog.String("interaction_guild_id", ic.GuildID),
		)
	}
}

func (h *CommandsHandler) handleCommand(
	cmdsFunc map[CommandName]CommandFunc,
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
) {
	var userID string
	if ic.User != nil {
		userID = ic.User.ID
	} else {
		userID = ic.Member.User.ID
	}

	data := ic.ApplicationCommandData()

	log := h.logger.With(
		slog.String("command_data_id", data.ID),
		slog.String("command_data_name", data.Name),
		slog.String("interaction_user_id", userID),
		slog.String("interaction_guild_id", ic.GuildID),
	)

	if hf, ok := cmdsFunc[data.Name]; ok {
		log.Debug("Handling application command.")

		if err := hf(s, ic); err != nil {
			log.Error("Failed to run command, error returned.",
				slog.String("error", err.Error()),
			)
		} else {
			log.Debug("Command ran successfully.")
		}

	} else {
		log.Error("Application command interaction created without having a handler.")
	}
}
