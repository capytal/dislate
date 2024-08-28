package errors

import (
	"fmt"
	"log/slog"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
)

type Error interface {
	Log(*slog.Logger)
	Reply(*dgo.Session, *dgo.Message)
	LogReply(*slog.Logger, *dgo.Session, *dgo.Message)
	Error() string
}

type defaultError struct {
	err  string
	args []slog.Attr
}

func NewError(err string, args ...slog.Attr) defaultError {
	return defaultError{err, args}
}

func New(err string, args ...slog.Attr) defaultError {
	return NewError(err, args...)
}

func (err defaultError) Log(l *slog.Logger) {
	args := make([]any, len(err.args))
	for i, a := range err.args {
		args[i] = any(a)
	}
	l.Error(err.err, args...)
}

func (err defaultError) Reply(s *dgo.Session, m *dgo.Message) {
	_, erro := s.ChannelMessageSendReply(
		m.ChannelID,
		fmt.Sprintf("Error: %s\nSee logs for more details", err.err),
		m.Reference(),
	)
	if erro != nil {
		_, _ = s.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("Failed to send error message (somehow), due to:\n%s", erro.Error()),
			m.Reference(),
		)
	}
}

func (err defaultError) Send(s *dgo.Session, channelID string) {
	_, erro := s.ChannelMessageSend(
		channelID,
		fmt.Sprintf("Error: %s\nSee logs for more details", err.err),
	)
	if erro != nil {
		_, _ = s.ChannelMessageSend(
			channelID,
			fmt.Sprintf("Failed to send error message (somehow), due to:\n%s", erro.Error()),
		)
	}
}

func (err defaultError) LogReply(l *slog.Logger, s *dgo.Session, m *dgo.Message) {
	err.Reply(s, m)
	err.Log(l)
}

func (err defaultError) LogSend(l *slog.Logger, s *dgo.Session, channelID string) {
	err.Send(s, channelID)
	err.Log(l)
}

func (err defaultError) Error() string {
	s := make([]string, len(err.args))
	for i, a := range err.args {
		s[i] = fmt.Sprintf("%s=%s", a.Key, a.Value)
	}
	return fmt.Sprintf("%s\n%s", err.err, strings.Join(s, " "))
}

type ErrDatabase struct {
	defaultError
}

func NewErrDatabase(args ...slog.Attr) ErrDatabase {
	return ErrDatabase{defaultError{
		"Error while trying to talk to the database.",
		args,
	}}
}

type ErrUserWebhook struct {
	defaultError
}

func NewErrUserWebhook(args ...slog.Attr) ErrUserWebhook {
	return ErrUserWebhook{defaultError{
		"Error while trying to access/execute the user webhook",
		args,
	}}
}
