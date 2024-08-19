package commands

import (
	"errors"
	"strings"

	"dislate/internals/guilddb"
	gdb "dislate/internals/guilddb"
	"dislate/internals/translator/lang"

	dgo "github.com/bwmarrin/discordgo"
)


type ManageChannel struct {
	db gdb.GuildDB
}

func NewManageChannel(db gdb.GuildDB) ManageChannel {
	return ManageChannel{db}
}
func (c ManageChannel) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionManageChannels

	return &dgo.ApplicationCommand{
		Name:                     "channel",
		Description:              "Manages a channel options",
		DefaultMemberPermissions: &permissions,
	}
}
func (c ManageChannel) Subcommands() []Command {
	return []Command{ChannelsInfo(c)}
}
func (c ManageChannel) Handle(s *dgo.Session, i *dgo.InteractionCreate) error {
	return nil
}

type ChannelsInfo struct {
	db gdb.GuildDB
}

func (c ChannelsInfo) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionManageChannels

	return &dgo.ApplicationCommand{
		Name:                     "info",
		Description:              "Get information about a channel",
		DefaultMemberPermissions: &permissions,
		Options: []*dgo.ApplicationCommandOption{{
			Type:        dgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "The channel to manage",
			ChannelTypes: []dgo.ChannelType{
				dgo.ChannelTypeGuildText,
			},
		}},
	}
}
func (c ChannelsInfo) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	opts := getOptions(ic)

	var err error

	var dch *dgo.Channel
	if c, ok := opts["channel"]; ok {
		dch = c.ChannelValue(s)
	} else {
		dch, err = s.Channel(ic.ChannelID)
		if err != nil {
			return err
		}
	}

	ch, err := getChannel(c.db, dch.GuildID, dch.ID)
	if err != nil {
		return err
	}

	info, err := getChannelInfo(c.db, ch)
	if err != nil {
		return err
	}

	err = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Embeds: []*dgo.MessageEmbed{info},
			Flags:  dgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
func (c ChannelsInfo) Subcommands() []Command {
	return []Command{}
}

func getChannel(db guilddb.GuildDB, guildID, channelID string) (gdb.Channel, error) {
	ch, err := db.Channel(guildID, channelID)
	if err != nil && errors.Is(err, gdb.ErrNotFound) {
		if err := db.ChannelInsert(gdb.NewChannel(guildID, channelID, lang.EN)); err != nil {
			return gdb.Channel{}, err
		}
		ch, err = db.Channel(guildID, channelID)
		if err != nil {
			return gdb.Channel{}, err
		}
	} else if err != nil {
		return gdb.Channel{}, err
	}

	return ch, nil
}

func getChannelInfo(db guilddb.GuildDB, ch gdb.Channel) (*dgo.MessageEmbed, error) {
	group, err := db.ChannelGroup(ch.GuildID, ch.ID)
	if err != nil && !errors.Is(err, gdb.ErrNotFound) {
		return nil, err
	}

	g := make([]string, len(group))
	for i, gi := range group {
		g[i] = "<#" + gi.ID + ">"
	}

	return &dgo.MessageEmbed{Title: "Channel Information",
		Fields: []*dgo.MessageEmbedField{
			{Name: "ID", Value: ch.ID, Inline: true},
			{Name: "Language", Value: string(ch.Language), Inline: true},
			{Name: "Linked Channels", Value: strings.Join(g, ", "), Inline: true},
		},
	}, nil
}
