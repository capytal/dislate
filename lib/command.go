package bot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	ApplicationCommand() *discordgo.ApplicationCommand
	Validate() (bool, error)
}

type Handler[CTX any] interface {
	Handle(ctx CTX) error
}

type HandlerFunc[CTX any] func(ctx CTX) error

func (h HandlerFunc[CTX]) Handle(ctx CTX) error {
	return h(ctx)
}

type Context struct {
	discordgo.Interaction
}

type ChatCommandContext struct {
	Context
	Options ChatCommandContextOptions
}

var (
	ErrChatCommandOptionNotExists   = errors.New("Chat command option does not exist")
	ErrChatCommandOptionInvalidType = errors.New("Chat command option is not of the type requested")
)

type ChatCommandContextOptions map[string]ChatCommandOption

func get[T ChatCommandOption](opts ChatCommandContextOptions, key string) (T, error) {
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

func (opts ChatCommandContextOptions) GetAttachement(
	key string,
) (*ChatCommandAttachmentOption, error) {
	return get[*ChatCommandAttachmentOption](opts, key)
}

func (opts ChatCommandContextOptions) GetBoolean(key string) (*ChatCommandBooleanOption, error) {
	return get[*ChatCommandBooleanOption](opts, key)
}

func (opts ChatCommandContextOptions) GetChannel(key string) (*ChatCommandChannelOption, error) {
	return get[*ChatCommandChannelOption](opts, key)
}

func (opts ChatCommandContextOptions) GetInteger(key string) (*ChatCommandIntegerOption, error) {
	return get[*ChatCommandIntegerOption](opts, key)
}

func (opts ChatCommandContextOptions) GetMentionable(
	key string,
) (*ChatCommandMentionableOption, error) {
	return get[*ChatCommandMentionableOption](opts, key)
}

func (opts ChatCommandContextOptions) GetNumber(key string) (*ChatCommandNumberOption, error) {
	return get[*ChatCommandNumberOption](opts, key)
}

func (opts ChatCommandContextOptions) GetRole(key string) (*ChatCommandRoleOption, error) {
	return get[*ChatCommandRoleOption](opts, key)
}

func (opts ChatCommandContextOptions) GetString(key string) (*ChatCommandStringOption, error) {
	return get[*ChatCommandStringOption](opts, key)
}

func (opts ChatCommandContextOptions) GetUser(key string) (*ChatCommandUserOption, error) {
	return get[*ChatCommandUserOption](opts, key)
}
