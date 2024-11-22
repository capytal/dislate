package bot

import (
	"errors"
	"slices"

	"github.com/bwmarrin/discordgo"
)

type ChatCommand struct {
	Name                     string
	NameLocalizations        *map[discordgo.Locale]string
	DefaultMemberPermissions *int64
	NSFW                     *bool
	Description              string
	DescriptionLocalizations *map[discordgo.Locale]string
	Options                  []ChatCommandOption
	Handler                  Handler
}

func (c *ChatCommand) ApplicationCommand() *discordgo.ApplicationCommand {
	requiredOpts := []*discordgo.ApplicationCommandOption{}
	opts := []*discordgo.ApplicationCommandOption{}

	for _, o := range c.Options {
		opt := o.ApplicationCommandOption()
		if opt.Required {
			requiredOpts = append(requiredOpts, opt)
		} else {
			opts = append(opts, opt)
		}
	}

	return &discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     c.Name,
		NameLocalizations:        c.NameLocalizations,
		DefaultMemberPermissions: c.DefaultMemberPermissions,
		NSFW:                     c.NSFW,
		Description:              c.Description,
		DescriptionLocalizations: c.DescriptionLocalizations,
		Options:                  slices.Concat(requiredOpts, opts),
	}
}

func (c *ChatCommand) Validate() (bool, error) {
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

type ChatCommandOption interface {
	ApplicationCommandOption() *discordgo.ApplicationCommandOption
	Validate() (bool, error)
}
