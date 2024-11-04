package commands

import (
	dgo "github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *dgo.ApplicationCommand
	Handle(s *dgo.Session, i *dgo.InteractionCreate) error
	Subcommands() []Command
	Components() []Component
}

type Component interface {
	Info() dgo.MessageComponent
	Handle(s *dgo.Session, i *dgo.InteractionCreate) error
}

func getOptions(
	opts []*dgo.ApplicationCommandInteractionDataOption,
) map[string]*dgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*dgo.ApplicationCommandInteractionDataOption, len(opts))

	for _, opt := range opts {
		if opt.Type == dgo.ApplicationCommandOptionSubCommand {
			return getOptions(opt.Options)
		} else {
			m[opt.Name] = opt
		}
	}

	return m
}
