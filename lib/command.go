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
}

}
