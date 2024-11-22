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
	Handler                  Handler[MessageCommandCtx]
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

func (c *MessageCommand) Validate() error {
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

type MessageCommandCtx struct {
	Ctx
}
