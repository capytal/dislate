package bot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type UserCommand struct {
	Name                     string
	NameLocalizations        *map[discordgo.Locale]string
	DefaultMemberPermissions *int64
	NSFW                     *bool
	Description              string
	DescriptionLocalizations *map[discordgo.Locale]string
	Handler                  Handler[UserCommandCtx]
}

func (c *UserCommand) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:                     discordgo.UserApplicationCommand,
		Name:                     c.Name,
		NameLocalizations:        c.NameLocalizations,
		DefaultMemberPermissions: c.DefaultMemberPermissions,
		NSFW:                     c.NSFW,
		Description:              c.Description,
		DescriptionLocalizations: c.DescriptionLocalizations,
	}
}

func (c *UserCommand) Validate() error {
	switch {
	case c.Name == "":
		return errors.New("required field \"Name\" is empty")
	case c.Description == "":
		return errors.New("required field \"Description\" is empty")
	case c.Handler == nil:
		return errors.New("required field \"Handler\" is empty")
	}
	return nil
}

type UserCommandCtx struct {
	Ctx
}
