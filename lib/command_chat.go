package bot

import (
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrChatCommandOptionNotExists   = errors.New("chat command option does not exist")
	ErrChatCommandOptionInvalidType = errors.New("chat command option is not of the type requested")
)

type ChatCommand struct {
	Name                     string
	NameLocalizations        *map[discordgo.Locale]string
	DefaultMemberPermissions *int64
	NSFW                     *bool
	Description              string
	DescriptionLocalizations *map[discordgo.Locale]string
	Options                  []ChatCommandOption
	Handler                  Handler[ChatCommandCtx]
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

func (c *ChatCommand) Validate() error {
	switch {
	case c.Name == "":
		return errors.New("required field \"Name\" is empty")
	case c.Description == "":
		return errors.New("required field \"Description\" is empty")
	case c.Handler == nil:
		return errors.New("required field \"Handler\" is empty")
	}

	for _, opt := range c.Options {
		if err := opt.Validate(); err != nil {
			return errors.Join(
				fmt.Errorf("option %q is not valid", opt.ApplicationCommandOption().Name),
				err,
			)
		}
	}

	return nil
}

type ChatCommandCtx struct {
	Ctx
	Options ChatCommandCtxOptions
}

type ChatCommandCtxOptions map[string]ChatCommandOption

func (opts ChatCommandCtxOptions) GetAttachement(
	key string,
) (*ChatCommandAttachmentOption, error) {
	return get[*ChatCommandAttachmentOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetBoolean(key string) (*ChatCommandBooleanOption, error) {
	return get[*ChatCommandBooleanOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetChannel(key string) (*ChatCommandChannelOption, error) {
	return get[*ChatCommandChannelOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetInteger(key string) (*ChatCommandIntegerOption, error) {
	return get[*ChatCommandIntegerOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetMentionable(
	key string,
) (*ChatCommandMentionableOption, error) {
	return get[*ChatCommandMentionableOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetNumber(key string) (*ChatCommandNumberOption, error) {
	return get[*ChatCommandNumberOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetRole(key string) (*ChatCommandRoleOption, error) {
	return get[*ChatCommandRoleOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetString(key string) (*ChatCommandStringOption, error) {
	return get[*ChatCommandStringOption](opts, key)
}

func (opts ChatCommandCtxOptions) GetUser(key string) (*ChatCommandUserOption, error) {
	return get[*ChatCommandUserOption](opts, key)
}

type ChatCommandOption interface {
	ApplicationCommandOption() *discordgo.ApplicationCommandOption
	Validate() error
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

func (o *ChatCommandAttachmentOption) Validate() error {
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

func (o *ChatCommandBooleanOption) Validate() error {
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

func (o *ChatCommandChannelOption) Validate() error {
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

func (o *ChatCommandIntegerOption) Validate() error {
	for _, c := range o.Choices {
		if c.Value < o.MinValue {
			return fmt.Errorf(
				"choice %q has value (%v) smaller than allowed by field \"MinValue\" (%v)",
				c.Name,
				c.Value,
				o.MinValue,
			)
		} else if c.Value > o.MaxValue {
			return fmt.Errorf(
				"choice %q has value (%v) bigger than allowed by field \"MaxValue\" (%v)",
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

func (o *ChatCommandMentionableOption) Validate() error {
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

func (o *ChatCommandNumberOption) Validate() error {
	for _, c := range o.Choices {
		if c.Value < o.MinValue {
			return fmt.Errorf(
				"choice %q has value (%v) smaller than allowed by field \"MinValue\" (%v)",
				c.Name,
				c.Value,
				o.MinValue,
			)
		} else if c.Value > o.MaxValue {
			return fmt.Errorf(
				"choice %q has value (%v) bigger than allowed by field \"MaxValue\" (%v)",
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

func (o *ChatCommandRoleOption) Validate() error {
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

func (o *ChatCommandStringOption) Validate() error {
	if o.MinLength > 6000 {
		return errors.New(
			"field \"MinLength\" has value that exceeds the allowed limit of 6000",
		)
	} else if o.MaxLength > 6000 {
		return errors.New(
			"field \"MaxLength\" has value that exceeds the allowed limit of 6000",
		)
	}

	for _, c := range o.Choices {
		l := utf8.RuneCountInString(c.Value)
		if l < o.MinLength {
			return fmt.Errorf(
				"choice %q has value (%q) with length (%v) smaller than allowed by \"MinLength\" field (%v)",
				c.Name,
				l,
				c.Value,
				o.MinLength,
			)
		} else if l > o.MaxLength {
			return fmt.Errorf(
				"choice %q has value (%q) with length (%v) bigger than allowed by \"MaxLength\" field (%v)",
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

func (o *ChatCommandUserOption) Validate() error {
	return validateOption(o)
}

type (
	optionTypes                            interface{ string | int | float64 }
	ChatCommandOptionChoice[T optionTypes] struct {
		Name              string
		NameLocalizations map[discordgo.Locale]string
		Value             T
	}
)

func get[T ChatCommandOption](opts ChatCommandCtxOptions, key string) (T, error) {
	var v T

	mv, ok := opts[key]
	if !ok {
		return v, ErrChatCommandOptionNotExists
	}

	if v, ok := mv.(T); !ok {
		return v, ErrChatCommandOptionInvalidType
	} else {
		return v, nil
	}
}

func validateOption(opt interface {
	ApplicationCommandOption() *discordgo.ApplicationCommandOption
},
) error {
	o := opt.ApplicationCommandOption()

	switch {
	case o.Name == "":
		return errors.New("required field \"Name\" is empty")
	case o.Description == "":
		return errors.New("required field  \"Description\" is empty")
	case len(o.Choices) > 0 && o.Autocomplete:
		return errors.New(
			"mutually exclusive fields \"Choices\" and \"Autocomplete\" are setted",
		)
	}

	return nil
}
