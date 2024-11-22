package bot

import (
	"errors"
	"fmt"
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

	for _, opt := range c.Options {
		if ok, err := opt.Validate(); !ok {
			return false, errors.Join(
				fmt.Errorf("Option %q is not valid", opt.ApplicationCommandOption().Name),
				err,
			)
		}
	}

	return true, nil
}

type ChatCommandOption interface {
	ApplicationCommandOption() *discordgo.ApplicationCommandOption
	Validate() (bool, error)
}

type ChatCommandStringOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandStringOptionChoice
	MinLength                int
	MaxLength                int
}

type ChatCommandStringOptionChoice struct {
	Name              string
	NameLocalizations map[discordgo.Locale]string
	Value             string
}

func (o *ChatCommandStringOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionString,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		MinLength:                &o.MinLength,
		MaxLength:                o.MaxLength,
		Choices:                  choices,
	}
}

func (o *ChatCommandStringOption) Validate() (bool, error) {
	if o.MinLength > 6000 {
		return false, errors.New(
			"Property \"MinLength\" has value that exceeds the allowed limit of 6000",
		)
	} else if o.MaxLength > 6000 {
		return false, errors.New(
			"Property \"MaxLength\" has value that exceeds the allowed limit of 6000",
		)
	}

	return validateOption(o)
}

type ChatCommandIntegerOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandIntegerOptionChoice
	MinValue                 int
	MaxValue                 int
}

type ChatCommandIntegerOptionChoice struct {
	Name              string
	NameLocalizations map[discordgo.Locale]string
	Value             int
}

func (o *ChatCommandIntegerOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	minValue := float64(o.MinValue)

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionInteger,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		MinValue:                 &minValue,
		MaxValue:                 float64(o.MaxValue),
		Choices:                  choices,
	}
}

func (o *ChatCommandIntegerOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandBooleanOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandBooleanOptionChoice
}

type ChatCommandBooleanOptionChoice struct {
	Name              string
	NameLocalizations map[discordgo.Locale]string
	Value             bool
}

func (o *ChatCommandBooleanOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionBoolean,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		Choices:                  choices,
	}
}

func (o *ChatCommandBooleanOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandUserOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandUserOptionChoice
}

type ChatCommandUserOptionChoice = ChatCommandStringOptionChoice

func (o *ChatCommandUserOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionUser,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		Choices:                  choices,
	}
}

func (o *ChatCommandUserOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandChannelOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandChannelOptionChoice
}

type ChatCommandChannelOptionChoice = ChatCommandStringOptionChoice

func (o *ChatCommandChannelOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionChannel,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		Choices:                  choices,
	}
}

func (o *ChatCommandChannelOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandRoleOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandRoleOptionChoice
}

type ChatCommandRoleOptionChoice = ChatCommandStringOptionChoice

func (o *ChatCommandRoleOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionRole,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		Choices:                  choices,
	}
}

func (o *ChatCommandRoleOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandMentionableOption struct {
	Name                     string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandMentionableOptionChoice
}

type ChatCommandMentionableOptionChoice = ChatCommandStringOptionChoice

func (o *ChatCommandMentionableOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionMentionable,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		Choices:                  choices,
	}
}

func (o *ChatCommandMentionableOption) Validate() (bool, error) {
	return validateOption(o)
}

func validateOption(opt interface {
	ApplicationCommandOption() *discordgo.ApplicationCommandOption
},
) (bool, error) {
	o := opt.ApplicationCommandOption()

	switch {
	case o.Name == "":
		return false, errors.New("Required property \"Name\" is empty")
	case o.Description == "":
		return false, errors.New("Required property \"Description\" is empty")
	}

	return true, nil
}
