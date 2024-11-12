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

func (h *CommandsHandler) RegisterCommands(
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

func equalCommand(left, right *discordgo.ApplicationCommand) (bool, error) {
	switch true {
	case left.Type != right.Type:
		return false, fmt.Errorf("Type is not equal. Left: %#v Right: %#v", left.Type, right.Type)

	case left.Name != right.Name:
		return false, fmt.Errorf("Name is not equal. Left: %#v Right: %#v", left.Name, right.Name)

	case (left.NameLocalizations != nil || right.NameLocalizations != nil) &&
		!reflect.DeepEqual(left.NameLocalizations, right.NameLocalizations):
		return false, fmt.Errorf(
			"NameLocalizations is not equal. Left: %#v Right: %#v",
			left.NameLocalizations,
			right.NameLocalizations,
		)

	case (left.DefaultMemberPermissions != nil || right.DefaultMemberPermissions != nil) &&
		!reflect.DeepEqual(left.DefaultMemberPermissions, right.DefaultMemberPermissions):
		return false, fmt.Errorf(
			"DefaultMemberPermissions is not equal. Left: %#v Right: %#v",
			left.DefaultMemberPermissions,
			right.DefaultMemberPermissions,
		)

	case (left.DMPermission != nil || right.DMPermission != nil) &&
		!reflect.DeepEqual(left.DMPermission, right.DMPermission):
		return false, fmt.Errorf(
			"DMPermission is not equal. Left: %#v Right: %#v",
			left.DMPermission,
			right.DMPermission,
		)

	case (left.NSFW != nil || right.NSFW != nil) && !reflect.DeepEqual(left.NSFW, right.NSFW):
		return false, fmt.Errorf("VALUE is not equal. Left: %#v Right: %#v", left.NSFW, right.NSFW)

	case (left.Description != "" || right.Description != "") && left.Description != right.Description:
		return false, fmt.Errorf(
			"Description is not equal. Left: %#v Right: %#v",
			left.Description,
			right.Description,
		)

	case (left.DescriptionLocalizations != nil || right.DescriptionLocalizations != nil) &&
		!reflect.DeepEqual(left.DescriptionLocalizations, right.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Left: %#v Right: %#v",
			left.DescriptionLocalizations,
			right.DescriptionLocalizations,
		)

	case len(left.Options) != len(right.Options):
		return false, fmt.Errorf(
			"Options is not equal. Left: %#v Right: %#v",
			left.Options,
			right.Options,
		)

	case len(left.Options) > 0 && len(right.Options) > 0:
		for i, o := range left.Options {
			if ok, err := equalCommandOption(o, right.Options[i]); !ok {
				return ok, errors.Join(fmt.Errorf("Option element of index %v has difference", err))
			}
		}
	}

	return true, nil
}

func equalCommandOption(left, right *discordgo.ApplicationCommandOption) (bool, error) {
	switch true {
	case left.Type != right.Type:
		return false, fmt.Errorf("Type is not equal. Left: %#v Right: %#v", left.Type, right.Type)

	case left.Name != right.Name:
		return false, fmt.Errorf("Name is not equal. Left: %#v Right: %#v", left.Name, right.Name)

	case left.Description != right.Description:
		return false, fmt.Errorf(
			"Description is not equal. Left: %#v Right: %#v",
			left.Description,
			right.Description,
		)

	case (left.DescriptionLocalizations != nil || right.DescriptionLocalizations != nil) &&
		!reflect.DeepEqual(left.DescriptionLocalizations, right.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Left: %#v Right: %#v",
			left.DescriptionLocalizations,
			right.DescriptionLocalizations,
		)

	case !reflect.DeepEqual(left.ChannelTypes, right.ChannelTypes):
		return false, fmt.Errorf(
			"ChannelTypes is not equal. Left: %#v Right: %#v",
			left.ChannelTypes,
			right.ChannelTypes,
		)

	case left.Required != right.Required:
		return false, fmt.Errorf(
			"Required is not equal. Left: %#v Right: %#v",
			left.Required,
			right.Required,
		)

	case !reflect.DeepEqual(left.Choices, right.Choices):
		return false, fmt.Errorf(
			"Choices is not equal. Left: %#v Right: %#v",
			left.Choices,
			right.Choices,
		)

	case (left.MinValue != nil || right.MinValue != nil) &&
		!reflect.DeepEqual(left.MinValue, right.MinValue):
		return false, fmt.Errorf(
			"MinValue is not equal. Left: %#v Right: %#v",
			left.MinValue,
			right.MinValue,
		)

	case (left.MaxValue != 0 || right.MaxValue != 0) && left.MaxValue != right.MaxValue:
		return false, fmt.Errorf(
			"MaxValue is not equal. Left: %#v Right: %#v",
			left.MaxValue,
			right.MaxValue,
		)

	case (left.MinLength != nil || right.MinLength != nil) &&
		!reflect.DeepEqual(left.MinLength, right.MinLength):
		return false, fmt.Errorf(
			"MinLength is not equal. Left: %#v Right: %#v",
			left.MinLength,
			right.MinLength,
		)

	case (left.MaxLength != 0 || right.MaxLength != 0) && left.MaxLength != right.MaxLength:
		return false, fmt.Errorf(
			"MaxLength is not equal. Left: %#v Right: %#v",
			left.MaxLength,
			right.MaxLength,
		)
	}

	return true, nil
}
