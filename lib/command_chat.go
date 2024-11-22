package bot

import (
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"

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
	Handler                  Handler[ChatCommandContext]
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

type ChatCommandAttachmentOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandAttachmentOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionAttachment,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandAttachmentOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandBooleanOption struct {
	Name                     string
	Value                    bool
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandBooleanOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionBoolean,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandBooleanOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandChannelOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandChannelOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionChannel,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandChannelOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandIntegerOption struct {
	Name                     string
	Value                    int
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandOptionChoice[int]
	MinValue                 int
	MaxValue                 int
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
	for _, c := range o.Choices {
		if c.Value < o.MinValue {
			return false, fmt.Errorf(
				"Choice %q has value (%v) smaller than allowed by property \"MinValue\" (%v)",
				c.Name,
				c.Value,
				o.MinValue,
			)
		} else if c.Value > o.MaxValue {
			return false, fmt.Errorf(
				"Choice %q has value (%v) bigger than allowed by property \"MaxValue\" (%v)",
				c.Name,
				c.Value,
				o.MaxValue,
			)
		}
	}

	return validateOption(o)
}

type ChatCommandMentionableOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandMentionableOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionMentionable,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandMentionableOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandNumberOption struct {
	Name                     string
	Value                    float64
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandOptionChoice[float64]
	MinValue                 float64
	MaxValue                 float64
}

func (o *ChatCommandNumberOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(o.Choices))
	for i, v := range o.Choices {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              v.Name,
			NameLocalizations: v.NameLocalizations,
			Value:             any(v.Value),
		}
	}

	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionNumber,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
		Autocomplete:             o.Autocomplete,
		MinValue:                 &o.MinValue,
		MaxValue:                 o.MaxValue,
		Choices:                  choices,
	}
}

func (o *ChatCommandNumberOption) Validate() (bool, error) {
	for _, c := range o.Choices {
		if c.Value < o.MinValue {
			return false, fmt.Errorf(
				"Choice %q has value (%v) smaller than allowed by property \"MinValue\" (%v)",
				c.Name,
				c.Value,
				o.MinValue,
			)
		} else if c.Value > o.MaxValue {
			return false, fmt.Errorf(
				"Choice %q has value (%v) bigger than allowed by property \"MaxValue\" (%v)",
				c.Name,
				c.Value,
				o.MaxValue,
			)
		}
	}

	return validateOption(o)
}

type ChatCommandRoleOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandRoleOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionRole,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandRoleOption) Validate() (bool, error) {
	return validateOption(o)
}

type ChatCommandStringOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
	Autocomplete             bool
	Choices                  []*ChatCommandOptionChoice[string]
	MinLength                int
	MaxLength                int
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

	for _, c := range o.Choices {
		l := utf8.RuneCountInString(c.Value)
		if l < o.MinLength {
			return false, fmt.Errorf(
				"Choice %q has value (%q) with length (%v) smaller than allowed by \"MinLength\" property (%v)",
				c.Name,
				l,
				c.Value,
				o.MinLength,
			)
		} else if l > o.MaxLength {
			return false, fmt.Errorf(
				"Choice %q has value (%q) with length (%v) bigger than allowed by \"MaxLength\" property (%v)",
				c.Name,
				l,
				c.Value,
				o.MaxLength,
			)
		}
	}

	return validateOption(o)
}

type ChatCommandUserOption struct {
	Name                     string
	Value                    string
	NameLocalizations        map[discordgo.Locale]string
	Description              string
	DescriptionLocalizations map[discordgo.Locale]string
	Required                 bool
}

func (o *ChatCommandUserOption) ApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:                     discordgo.ApplicationCommandOptionUser,
		Name:                     o.Name,
		NameLocalizations:        o.NameLocalizations,
		Description:              o.Description,
		DescriptionLocalizations: o.DescriptionLocalizations,
		Required:                 o.Required,
	}
}

func (o *ChatCommandUserOption) Validate() (bool, error) {
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
	case len(o.Choices) > 0 && o.Autocomplete:
		return false, errors.New(
			"Mutually exclusive properties \"Choices\" and \"Autocomplete\" are setted",
		)
	}

	return true, nil
}

type (
	optionTypes                            interface{ string | int | float64 }
	ChatCommandOptionChoice[T optionTypes] struct {
		Name              string
		NameLocalizations map[discordgo.Locale]string
		Value             T
	}
)
