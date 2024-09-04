package commands

import (
	"errors"
	"fmt"
	"strings"

	"dislate/internals/discord/bot/gconf"
	"dislate/internals/guilddb"
	"dislate/internals/translator/lang"

	gdb "dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type ManageChannel struct {
	db gconf.DB
}

func NewManageChannel(db gconf.DB) ManageChannel {
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
	return []Command{
		channelsInfo(c),
		channelsLink(c),
		channelsSetLang(c),
	}
}

func (c ManageChannel) Handle(s *dgo.Session, i *dgo.InteractionCreate) error {
	return nil
}

func (c ManageChannel) Components() []Component {
	return []Component{}
}

type channelsInfo struct {
	db gconf.DB
}

func (c channelsInfo) Info() *dgo.ApplicationCommand {
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
				dgo.ChannelTypeGuildForum,
				dgo.ChannelTypeGuildPublicThread,
				dgo.ChannelTypeGuildPrivateThread,
			},
		}},
	}
}

func (c channelsInfo) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	opts := getOptions(ic.ApplicationCommandData().Options)

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

func (c channelsInfo) Components() []Component {
	return []Component{}
}

func (c channelsInfo) Subcommands() []Command {
	return []Command{}
}

type channelsLink struct {
	db gconf.DB
}

func (c channelsLink) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionManageChannels

	return &dgo.ApplicationCommand{
		Name:                     "link",
		Description:              "Link two channels together",
		DefaultMemberPermissions: &permissions,
		Options: []*dgo.ApplicationCommandOption{{
			Type:        dgo.ApplicationCommandOptionChannel,
			Name:        "channel_one",
			Description: "The channel to link",
			Required:    true,
			ChannelTypes: []dgo.ChannelType{
				dgo.ChannelTypeGuildText,
				dgo.ChannelTypeGuildForum,
				dgo.ChannelTypeGuildPublicThread,
				dgo.ChannelTypeGuildPrivateThread,
			},
		}, {
			Type:        dgo.ApplicationCommandOptionChannel,
			Name:        "channel_two",
			Description: "The channel to link",
			ChannelTypes: []dgo.ChannelType{
				dgo.ChannelTypeGuildText,
				dgo.ChannelTypeGuildForum,
				dgo.ChannelTypeGuildPublicThread,
				dgo.ChannelTypeGuildPrivateThread,
			},
		}},
	}
}

func (c channelsLink) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	opts := getOptions(ic.ApplicationCommandData().Options)

	var err error
	var dch1, dch2 *dgo.Channel
	if c, ok := opts["channel_one"]; ok {
		dch1 = c.ChannelValue(s)
	} else {
		return errors.New("channel_one is required")
	}

	if c, ok := opts["channel_two"]; ok {
		dch2 = c.ChannelValue(s)
	} else {
		dch2, err = s.Channel(ic.ChannelID)
		if err != nil {
			return err
		}
	}

	if dch1.ID == dch2.ID {
		return errors.New("channel_one and channel_two must be different values")
	} else if dch1.Type != dch2.Type {
		return errors.New("channel_one and channel_two must be the same channel types")
	}

	ch1, err := getChannel(c.db, dch1.GuildID, dch1.ID)
	if err != nil {
		return err
	}
	ch2, err := getChannel(c.db, dch2.GuildID, dch2.ID)
	if err != nil {
		return err
	}

	var cb1, cb2 guilddb.ChannelGroup

	cb1, err = c.db.ChannelGroup(ch1.GuildID, ch1.ID)
	if err != nil && !errors.Is(err, guilddb.ErrNotFound) {
		return err
	}
	cb2, err = c.db.ChannelGroup(ch2.GuildID, ch2.ID)
	if err != nil && !errors.Is(err, guilddb.ErrNotFound) {
		return err
	}

	if len(cb1) > 0 && len(cb2) > 0 {
		return errors.New("both channels are already in a group")
	} else if len(cb1) > 0 {
		cb1 = append(cb1, ch2)
		err = c.db.ChannelGroupUpdate(cb1)
	} else if len(cb2) > 0 {
		cb2 = append(cb2, ch1)
		err = c.db.ChannelGroupUpdate(cb2)
	} else {
		err = c.db.ChannelGroupInsert(guilddb.ChannelGroup{ch1, ch2})
	}
	if err != nil {
		return err
	}
	err = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: fmt.Sprintf(
				"Linked channel %s (%s) and %s (%s)",
				dch1.Name, dch1.ID, dch2.Name, dch2.ID,
			),
			Flags: dgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c channelsLink) Components() []Component {
	return []Component{}
}

func (c channelsLink) Subcommands() []Command {
	return []Command{}
}

type channelsSetLang struct {
	db gconf.DB
}

func (c channelsSetLang) Info() *dgo.ApplicationCommand {
	var permissions int64 = dgo.PermissionManageChannels

	return &dgo.ApplicationCommand{
		Name:                     "set-lang",
		Description:              "Link two channels together",
		DefaultMemberPermissions: &permissions,
		Options: []*dgo.ApplicationCommandOption{{
			Type:        dgo.ApplicationCommandOptionString,
			Required:    true,
			Name:        "language",
			Description: "The new language",
			Choices: []*dgo.ApplicationCommandOptionChoice{
				{Name: "English (EN)", Value: lang.EN},
				{Name: "Portuguese (PT)", Value: lang.PT},
			},
		}, {
			Type:        dgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "The channel to change the language",
			ChannelTypes: []dgo.ChannelType{
				dgo.ChannelTypeGuildText,
				dgo.ChannelTypeGuildForum,
				dgo.ChannelTypeGuildPublicThread,
				dgo.ChannelTypeGuildPrivateThread,
			},
		}},
	}
}

func (c channelsSetLang) Handle(s *dgo.Session, ic *dgo.InteractionCreate) error {
	opts := getOptions(ic.ApplicationCommandData().Options)

	var err error
	var dch *dgo.Channel
	var l lang.Language

	if c, ok := opts["language"]; ok {
		switch c.StringValue() {
		case string(lang.PT):
			l = lang.PT
		default:
			l = lang.EN
		}
	} else {
		return errors.New("language is a required option")
	}

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

	ch.Language = l

	err = c.db.ChannelUpdate(ch)
	if err != nil {
		return err
	}

	err = s.InteractionRespond(ic.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: fmt.Sprintf(
				"Changed language of channel %s (%s) to %s",
				dch.Name, dch.ID, l,
			),
			Flags: dgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c channelsSetLang) Components() []Component {
	return []Component{}
}

func (c channelsSetLang) Subcommands() []Command {
	return []Command{}
}

func getChannel(db gconf.DB, guildID, channelID string) (gdb.Channel, error) {
	ch, err := db.Channel(guildID, channelID)
	if errors.Is(err, gdb.ErrNotFound) {
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

func getChannelInfo(db gconf.DB, ch gdb.Channel) (*dgo.MessageEmbed, error) {
	group, err := db.ChannelGroup(ch.GuildID, ch.ID)
	if err != nil && !errors.Is(err, gdb.ErrNotFound) {
		return nil, err
	}

	g := make([]string, len(group))
	for i, gi := range group {
		g[i] = "<#" + gi.ID + ">"
	}

	return &dgo.MessageEmbed{
		Title: "Channel Information",
		Fields: []*dgo.MessageEmbedField{
			{Name: "ID", Value: ch.ID, Inline: true},
			{Name: "Language", Value: string(ch.Language), Inline: true},
			{Name: "Linked Channels", Value: strings.Join(g, ", "), Inline: true},
		},
	}, nil
}
