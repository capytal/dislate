package commands

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *discordgo.ApplicationCommand
	Handle(s *discordgo.Session, ic *discordgo.InteractionCreate) error
}

type (
	CommandFunc = func(s *discordgo.Session, ic *discordgo.InteractionCreate) error
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
	commands []Command,
	guildID ...string,
) error {
	var GUILD_ID string
	if len(guildID) > 0 {
		GUILD_ID = guildID[0]
	}

	commandsMap, err := h.mapCommands(commands)
	if err != nil {
		return err
	}

	APP_ID := h.session.State.User.ID
	if APP_ID == "" {
		return errors.New("User ID is not set in session state")
	}

	registeredCommands, err := h.mapRegisteredCommmands(GUILD_ID)
	if err != nil {
		return err
	}

	if err := h.removeUnhandledCommands(commandsMap, registeredCommands, GUILD_ID); err != nil {
		return err
	}

	handleFuncs := make(map[CommandName]CommandFunc, len(commandsMap))

	for _, cmd := range commandsMap {
		var err error

		appCmd, isRegistered := registeredCommands[cmd.Info().Name]

		if !isRegistered {
			h.logger.Debug("Bot command is not registered in application, registering.",
				slog.String("command_name", cmd.Info().Name),
				slog.String("guild_id", GUILD_ID))

			appCmd, err = h.session.ApplicationCommandCreate(APP_ID, GUILD_ID, cmd.Info())
			if err != nil {
				return err
			}

		} else if y, err := equalToRegistered(cmd.Info(), appCmd); !y {
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

	h.session.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		h.handleInteraction(handleFuncs, s, ic)
	})

	return nil
}

func (h *CommandsHandler) mapCommands(commands []Command) (map[CommandName]Command, error) {
	m := make(map[CommandName]Command, len(commands))

	for _, c := range commands {
		if n := c.Info().Name; n != "" {
			m[c.Info().Name] = c
		} else {
			return m, fmt.Errorf("Command doesn't have a valid name!")
		}
	}

	return m, nil
}

func (h *CommandsHandler) mapRegisteredCommmands(
	guildID string,
) (map[CommandName]*discordgo.ApplicationCommand, error) {
	cmdMap := map[CommandName]*discordgo.ApplicationCommand{}

	registeredCommands, err := h.session.ApplicationCommands(h.session.State.User.ID, guildID)
	if err != nil {
		return cmdMap, err
	}
	for _, rc := range registeredCommands {
		cmdMap[rc.Name] = rc
	}

	return cmdMap, nil
}

func (h *CommandsHandler) removeUnhandledCommands(
	handledCommands map[CommandName]Command,
	registeredCommands map[CommandName]*discordgo.ApplicationCommand,
	guildID string,
) error {
	for _, cmd := range registeredCommands {
		if _, isHandled := handledCommands[cmd.Name]; !isHandled {
			h.logger.Debug("Registered command no longer is being handled, deleting.",
				slog.String("registered_command_name", cmd.Name),
				slog.String("registered_command_id", cmd.ID),
				slog.String("guild_id", guildID))

			err := h.session.ApplicationCommandDelete(h.session.State.User.ID, guildID, cmd.ID)
			if err != nil {
				return err
			}

			delete(registeredCommands, cmd.Name)
		}
	}

	return nil
}
