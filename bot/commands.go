package bot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type mockCommand struct{}

func (c *mockCommand) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "mock",
		Type:        discordgo.ChatApplicationCommand,
		Description: "this is a mock command used for testing",
		Options: []*discordgo.ApplicationCommandOption{{
			Name:        "message",
			Description: "The message to respond",
			Type:        discordgo.ApplicationCommandOptionString,
		}},
	}
}

func (c *mockCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		return errors.New("Option \"message\" is not defined")
	}

	text, ok := data.Options[0].Value.(string)
	if !ok {
		return errors.New("Failed to convert \"message\" into string")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Hello user! Your text is %q", text),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
