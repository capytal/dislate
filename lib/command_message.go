package bot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type MessageCommand struct {
	Name                     string
	NameLocalizations        *map[discordgo.Locale]string
	DefaultMemberPermissions *int64
	NSFW                     *bool
	Description              string
	DescriptionLocalizations *map[discordgo.Locale]string
	Handler                  Handler
}

func (c *MessageCommand) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:                     discordgo.MessageApplicationCommand,
		Name:                     c.Name,
		NameLocalizations:        c.NameLocalizations,
		DefaultMemberPermissions: c.DefaultMemberPermissions,
		NSFW:                     c.NSFW,
		Description:              c.Description,
		DescriptionLocalizations: c.DescriptionLocalizations,
	}
}

func (c *MessageCommand) Validate() (bool, error) {
	switch {
	case c.Name == "":
		return false, errors.New("Required property \"Name\" is empty")
	case c.Description == "":
		return false, errors.New("Required property \"Description\" is empty")
	case c.Handler == nil:
		return false, errors.New("Required property \"Handler\" is empty")
	}
	return true, nil
}
