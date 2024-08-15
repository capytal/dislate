package commands

import (
	"dislate/internals/translator"
	"dislate/internals/translator/lang"

	dgo "github.com/bwmarrin/discordgo"
)

type Command interface {
	Info() *dgo.ApplicationCommand
	Handle(s *dgo.Session, i *dgo.InteractionCreate)
}

type Test struct {
	translator translator.Translator
}

func NewTest(t translator.Translator) Test {
	return Test{t}
}
func (c Test) Info() *dgo.ApplicationCommand {
	return &dgo.ApplicationCommand{
		Name:        "test-command",
		Description: "This is a test command",
	}
}
func (c Test) Handle(s *dgo.Session, i *dgo.InteractionCreate) {
	txt, _ := c.translator.Translate(lang.EN, lang.PT, "Hello world!")

	_ = s.InteractionRespond(i.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: txt,
		},
	})
}
