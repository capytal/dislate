package commands

import (
	dgo "github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *dgo.ApplicationCommand
	Handle(s *dgo.Session, i *dgo.InteractionCreate) error
	Subcommands() []Command
}
func getOptions(i *dgo.InteractionCreate) map[string]*dgo.ApplicationCommandInteractionDataOption {
	opts := i.ApplicationCommandData().Options
	m := make(map[string]*dgo.ApplicationCommandInteractionDataOption, len(opts))

	for _, opt := range opts {
		m[opt.Name] = opt
	}

	return m
}

