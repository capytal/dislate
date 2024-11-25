package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type commandID struct {
	Name    string
	GuildID string
}

type sessionCommandID struct {
	ID      string
	GuildID string
}

type (
	chatCommandHandlers    map[sessionCommandID]Handler[ChatCommandCtx]
	messageCommandHandlers map[sessionCommandID]Handler[MessageCommandCtx]
	userCommandHandlers    map[sessionCommandID]Handler[UserCommandCtx]
)

func (b *Bot) MustHandleCommand(c Command, guildID ...string) {
	if err := b.HandleCommand(c, guildID...); err != nil {
		panic(err)
	}
}

func (b *Bot) HandleCommand(c Command, guildID ...string) error {
	var g string
	if len(guildID) > 0 {
		g = guildID[0]
	}

	if err := c.Validate(); err != nil {
		return fmt.Errorf("failed to handle command %q: %v", c.ApplicationCommand().Name, err)
	}

	sessionCmd, err := b.sessionCommand(c.ApplicationCommand().Name, g)
	if err != nil {
		return err
	}

	return b.updateCommand(c, sessionCmd, g)
}

func (b *Bot) updateCommand(
	botCmd Command,
	sessionCmd *discordgo.ApplicationCommand,
	guildID ...string,
) error {
	var err error

	var g string
	if len(guildID) > 0 {
		g = guildID[0]
	}

	if sessionCmd == nil {
		sessionCmd, err = b.session.ApplicationCommandCreate(
			b.session.State.User.ID,
			g,
			botCmd.ApplicationCommand(),
		)
		if err != nil {
			return err
		}
	} else {
		sessionCmd, err = b.session.ApplicationCommandEdit(
			b.session.State.User.ID,
			g,
			sessionCmd.ID,
			botCmd.ApplicationCommand(),
		)
		if err != nil {
			return err
		}
	}

	switch botCmd.ApplicationCommand().Type {
	case discordgo.ChatApplicationCommand:
		if c, ok := botCmd.(*ChatCommand); ok {
			b.chatCommandHandlers[sessionCommandID{sessionCmd.ID, g}] = c.Handler
		} else {
			return err
		}
	case discordgo.MessageApplicationCommand:
		if c, ok := botCmd.(*MessageCommand); ok {
			b.messageCommandHandlers[sessionCommandID{sessionCmd.ID, g}] = c.Handler
		} else {
			return err
		}
	case discordgo.UserApplicationCommand:
		if c, ok := botCmd.(*UserCommand); ok {
			b.userCommandHandlers[sessionCommandID{sessionCmd.ID, g}] = c.Handler
		} else {
			return err
		}
	}

	return nil
}

func (b *Bot) sessionCommand(
	name string,
	guildID ...string,
) (*discordgo.ApplicationCommand, error) {
	var g string
	if len(guildID) > 0 {
		g = guildID[0]
	}

	sessionCmds, err := b.sessionCommands(g)
	if err != nil {
		return nil, err
	}

	cmd, ok := sessionCmds[commandID{Name: name, GuildID: g}]
	if !ok {
		return nil, nil
	}

	return cmd, nil
}

func (b *Bot) sessionCommands(
	guildID ...string,
) (map[commandID]*discordgo.ApplicationCommand, error) {
	var g string
	if len(guildID) > 0 {
		g = guildID[0]
	}

	appCmds, err := b.session.ApplicationCommands(b.session.State.User.ID, g)
	if err != nil {
		return nil, err
	}

	cmds := make(map[commandID]*discordgo.ApplicationCommand, len(appCmds))
	for _, c := range appCmds {
		cmds[commandID{Name: c.Name, GuildID: g}] = c
	}

	return cmds, err
}
