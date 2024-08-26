package commands

import (
	"dislate/internals/discord/bot/gconf"
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
)

type ManageConfig struct {
	db gconf.DB
}

func NewMagageConfig(db gconf.DB) ManageConfig {
	return ManageConfig{db}
}
func (c ManageConfig) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionAdministrator

	return &dgo.ApplicationCommand{
		Name:                     "config",
		Description:              "Manages the guild's configuration",
		DefaultMemberPermissions: &permissions,
	}
}
func (c ManageConfig) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	return nil
}
func (c ManageConfig) Components() []Component {
	return []Component{}
}
func (c ManageConfig) Subcommands() []Command {
	return []Command{
		loggerConfigChannel(c),
	}
}

type loggerConfigChannel struct {
	db gconf.DB
}
func (c loggerConfigChannel) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionAdministrator
	return &dgo.ApplicationCommand{
		Name:                     "log-channel",
		Description:              "Change logging channel",
		DefaultMemberPermissions: &permissions,
		Options: []*dgo.ApplicationCommandOption{{
			Type: dgo.ApplicationCommandOptionChannel,
			Required: true,
			Name: "log-channel",
			Description: "The channel to send log messages and errors to",
			ChannelTypes: []dgo.ChannelType{
				dgo.ChannelTypeGuildText,
			},
		}},
	}
}
func (c loggerConfigChannel) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	opts := getOptions(ic.ApplicationCommandData().Options)

	var err error
	var dch *dgo.Channel
	if c, ok := opts["log-channel"]; ok {
		dch = c.ChannelValue(s)
	} else {
		dch, err = s.Channel(ic.ChannelID)
		if err != nil {
			return err
		}
	}

	guild, err := c.db.Guild(ic.GuildID)
	if err != nil {
		return err
	}

	conf := guild.Config
	conf.LoggingChannel = &dch.ID
	guild.Config = conf

	err = c.db.GuildUpdate(guild)
	if err != nil {
		return err
	}

	err = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: fmt.Sprintf("Logging channel changed to %s", *guild.Config.LoggingChannel),
			Flags: dgo.MessageFlagsEphemeral,
		},
	})

	return err
}
func (c loggerConfigChannel) Components() []Component {
	return []Component{}
}
func (c loggerConfigChannel) Subcommands() []Command {
	return []Command{}
}
