package bot

import "github.com/bwmarrin/discordgo"

type Command interface {
	ApplicationCommand() *discordgo.ApplicationCommand
	Validate() (bool, error)
}

type Handler interface {
	Handle(s *discordgo.Session, ic *discordgo.InteractionCreate) error
}

type HandlerFunc func(s *discordgo.Session, ic *discordgo.InteractionCreate) error

func (h HandlerFunc) Handle(
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
) error {
	return h(s, ic)
}

