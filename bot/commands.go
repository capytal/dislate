package bot

import (
	"errors"
	"fmt"

	"forge.capytal.company/capytal/dislate/commands"
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

func (c *mockCommand) Handle(
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) error {
	if len(data.Options) == 0 {
		return errors.New("Option \"message\" is not defined")
	}

	text, ok := data.Options[0].Value.(string)
	if !ok {
		return errors.New("Failed to convert \"message\" into string")
	}

	if _, err := s.ChannelMessageSendComplex(ic.ChannelID, &discordgo.MessageSend{
		Content: "test component",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					c.Components()[0].Info(),
				},
			},
		},
	}); err != nil {
		return err
	}

	return s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Hello user! Your text is %q", text),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *mockCommand) Components() []commands.Component {
	return []commands.Component{
		&mockButton{},
	}
}

type mockButton struct{}

func (c *mockButton) Info() discordgo.MessageComponent {
	return &discordgo.Button{
		Label:    "test button",
		CustomID: "mock-command-test-button",
	}
}

func (c *mockButton) Handle(
	s *discordgo.Session,
	ic *discordgo.InteractionCreate,
	data discordgo.MessageComponentInteractionData,
) error {
	return s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Button clicked!"),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
