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

func getOptions(i *dgo.InteractionCreate) map[string]*dgo.ApplicationCommandInteractionDataOption {
	opts := i.ApplicationCommandData().Options
	m := make(map[string]*dgo.ApplicationCommandInteractionDataOption, len(opts))

	for _, opt := range opts {
		m[opt.Name] = opt
	}

	return m
}

