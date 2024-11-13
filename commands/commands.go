package commands

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *discordgo.ApplicationCommand
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type (
	CommandFunc = func(s *discordgo.Session, i *discordgo.InteractionCreate)
	CommandName = string
	CommandId   = string
)

type CommandsHandler struct {
	logger  *slog.Logger
	session *discordgo.Session
}

func NewCommandsHandler(logger *slog.Logger, session *discordgo.Session) *CommandsHandler {
	return &CommandsHandler{logger, session}
}

func (h *CommandsHandler) UpdateCommands(
	commands map[CommandName]Command,
	guildID ...string,
) error {
	var GUILD_ID string
	if len(guildID) > 0 {
		GUILD_ID = guildID[0]
	}

	APP_ID := h.session.State.User.ID
	if APP_ID == "" {
		return errors.New("User ID is not set in session state")
	}

	REGISTERED_COMMANDS := map[CommandName]*discordgo.ApplicationCommand{}
	registeredCommands, err := h.session.ApplicationCommands(APP_ID, GUILD_ID)
	if err != nil {
		return err
	}
	for _, rc := range registeredCommands {
		REGISTERED_COMMANDS[rc.Name] = rc
	}

	for _, cmd := range REGISTERED_COMMANDS {
		if _, isHandled := commands[cmd.Name]; !isHandled {
			h.logger.Debug("Registered command no longer is being handled, deleting.",
				slog.String("registered_command_name", cmd.Name),
				slog.String("registered_command_id", cmd.ID),
				slog.String("guild_id", GUILD_ID))

			err = h.session.ApplicationCommandDelete(APP_ID, GUILD_ID, cmd.ID)
			if err != nil {
				return err
			}

			delete(REGISTERED_COMMANDS, cmd.Name)
		}
	}

	handleFuncs := map[CommandName]CommandFunc{}

	for _, cmd := range commands {
		var err error

		appCmd, isRegistered := REGISTERED_COMMANDS[cmd.Info().Name]

		if !isRegistered {
			h.logger.Debug("Bot command is not registered in application, registering.",
				slog.String("command_name", cmd.Info().Name),
				slog.String("guild_id", GUILD_ID))

			appCmd, err = h.session.ApplicationCommandCreate(APP_ID, GUILD_ID, cmd.Info())
			if err != nil {
				return err
			}

		} else if ok, err := equalCommand(cmd.Info(), appCmd); !ok {
			h.logger.Debug("Bot command and registered command are different, deleting registered command for updating.",
				slog.String("command_name", cmd.Info().Name),
				slog.String("registered_command_id", appCmd.ID),
				slog.String("registered_command_name", appCmd.Name),
				slog.String("guild_id", GUILD_ID),
				slog.String("difference", err.Error()))

			err = h.session.ApplicationCommandDelete(APP_ID, GUILD_ID, appCmd.ID)
			if err != nil {
				return err
			}

			appCmd, err = h.session.ApplicationCommandCreate(APP_ID, GUILD_ID, cmd.Info())
			if err != nil {
				return err
			}
		}

		handleFuncs[appCmd.Name] = cmd.Handle
	}

	h.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			if hf, ok := handleFuncs[data.Name]; ok {
				h.logger.Debug("Handling application command.",
					slog.String("command_data_id", data.ID),
					slog.String("command_data_name", data.Name),
					slog.String("interaction_user_id", i.User.ID),
					slog.String("interaction_guild_id", i.GuildID),
				)
				hf(s, i)

			} else {
				h.logger.Error("Application command interaction created without having a handler.",
					slog.String("command_data_id", data.ID),
					slog.String("command_data_name", data.Name),
					slog.String("interaction_user_id", i.User.ID),
					slog.String("interaction_guild_id", i.GuildID),
				)
			}
		} else {
			h.logger.Error("Application interaction created without being a command.",
				slog.String("interaction_id", i.ID),
				slog.String("interaction_type", i.Type.String()),
				slog.String("interaction_user_id", i.User.ID),
				slog.String("interaction_guild_id", i.GuildID),
			)
		}
	})

	return nil
}

