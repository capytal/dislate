package commands

import (
	"dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *dgo.ApplicationCommand
	Handle(s *dgo.Session, i *dgo.InteractionCreate) error
}

type ManageChannel struct {
	db guilddb.GuildDB
}

func NewManageChannel(db guilddb.GuildDB) ManageChannel {
	return ManageChannel{db}
}
func (c ManageChannel) Info() *dgo.ApplicationCommand {
	return &dgo.ApplicationCommand{
		Name:        "channel",
		Description: "Manages a channel options",
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
func (c ManageChannel) Handle(s *dgo.Session, i *dgo.InteractionCreate) error {
	err := s.InteractionRespond(i.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: "Hello world!",
			Flags:   dgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
