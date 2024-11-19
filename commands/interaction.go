package commands

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func (h *CommandsHandler) handleInteraction(
	cmdsHandlers map[commandName]commandHandlerFunc,
	compsHandlers map[componentCustomId]componentHandlerFunc,
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
		h.handleCommandInteraction(cmdsHandlers, s, ic)
	case discordgo.InteractionMessageComponent:
		h.handleComponentInteraction(compsHandlers, s, ic)
	default:
		h.logger.Error("Application interaction is not a supported type!",
			slog.String("interaction_id", ic.ID),
			slog.String("interaction_type", ic.Type.String()),
			slog.String("interaction_user_id", userID),
			slog.String("interaction_guild_id", ic.GuildID),
		)
	}
}

func (h *CommandsHandler) handleCommandInteraction(
	cmdsHandlers map[commandName]commandHandlerFunc,
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

	if hf, ok := cmdsHandlers[data.Name]; ok {
		log.Debug("Handling application command.")

		if err := hf(s, ic, data); err != nil {
			log.Error("Failed to run command, error returned.", slog.String("error", err.Error()))

			_, err = s.ChannelMessageSendComplex(ic.ChannelID, &discordgo.MessageSend{
				Content: fmt.Sprintf(
					"<@%s>\nFailed to run command! Error message:\n```\n%s\n```",
					userID,
					err.Error(),
				),
			})
			if err != nil {
				log.Error(
					"Failed to send error message explaining error... somehow.",
					slog.String("error", err.Error()),
				)
			}

		} else {
			log.Debug("Command ran successfully.")
		}

	} else {
		log.Error("Application command interaction created without having a handler.")
	}
}

func (h *CommandsHandler) handleComponentInteraction(
	compsHandlers map[componentCustomId]componentHandlerFunc,
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
) {
	var userID string
	if ic.User != nil {
		userID = ic.User.ID
	} else {
		userID = ic.Member.User.ID
	}

	data := ic.MessageComponentData()

	log := h.logger.With(
		slog.String("component_data_custom_id", data.CustomID),
		slog.String("component_data_type", data.Type().String()),
		slog.String("interaction_user_id", userID),
		slog.String("interaction_guild_id", ic.GuildID),
	)

	if hf, ok := compsHandlers[data.CustomID]; ok {
		log.Debug("Handling application component.")

		if err := hf(s, ic, data); err != nil {
			log.Error(
				"Failed to handle component, error returned.",
				slog.String("error", err.Error()),
			)

			_, err = s.ChannelMessageSendComplex(ic.ChannelID, &discordgo.MessageSend{
				Content: fmt.Sprintf(
					"<@%s>\nFailed to handle component! Error message:\n```\n%s\n```",
					userID,
					err.Error(),
				),
			})
			if err != nil {
				log.Error(
					"Failed to send error message explaining error... somehow.",
					slog.String("error", err.Error()),
				)
			}

		} else {
			log.Debug("Component handled successfully.")
		}

	} else {
		log.Error("Application component interaction created without having a handler.")
	}
}
