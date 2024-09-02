package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
)

type BotError interface {
	Error() string
}

type BotErrorHandler interface {
	Info() string
	Error() string
	Reply(s *dgo.Session, m *dgo.Message) BotErrorHandler
	Send(s *dgo.Session, channelID string) BotErrorHandler
	Log(l *slog.Logger) BotErrorHandler
}

type defaultErrHandler struct {
	BotError
}

func (err *defaultErrHandler) Error() string {
	return err.Error()
}

func (err *defaultErrHandler) Info() string {
	return err.Info()
}

func (err *defaultErrHandler) Reply(s *dgo.Session, m *dgo.Message) BotErrorHandler {
	if _, erro := s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference()); erro != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf(
				"Failed to reply message %s due to \"%s\" with error: %s.",
				m.ID,
				erro.Error(),
				err.Error(),
			),
		)
	}
	return err
}

func (err *defaultErrHandler) Send(s *dgo.Session, channelID string) BotErrorHandler {
	if _, erro := s.ChannelMessageSend(channelID, err.Error()); erro != nil {
		_, _ = s.ChannelMessageSend(
			channelID,
			fmt.Sprintf(
				"Failed to send error message due to \"%s\" with error: %s.",
				erro.Error(),
				err.Error(),
			),
		)
	}
	return err
}

func (err *defaultErrHandler) Log(l *slog.Logger) BotErrorHandler {
	l.Error(err.Error())
	return err
}

type EventError[E any] struct {
	data   map[string]any
	errors []error
}

func NewEventError[E any](data map[string]any, err ...error) *EventError[E] {
	return &EventError[E]{data, err}
}

func (h *EventError[E]) Wrap(err ...error) *EventError[E] {
	h.errors = append(h.errors, errors.Join(err...))
	return h
}

func (h *EventError[E]) Wrapf(format string, a ...any) *EventError[E] {
	h.errors = append(h.errors, fmt.Errorf(format, a...))
	return h
}

func (h *EventError[E]) AddData(key string, value any) *EventError[E] {
	h.data[key] = value
	return h
}

func (h *EventError[E]) Error() string {
	var ev E
	var name string
	if t := reflect.TypeOf(ev); t != nil {
		if n := t.Name(); n != "" {
			name = strings.ToUpper(n)
		} else {
			name = "UNAMED EVENT"
		}
	} else {
		name = "UNAMED EVENT"
	}
	err := errors.Join(h.errors...)
	return errors.Join(fmt.Errorf("Failed to process event %s", name), err).Error()
}

func (h *EventError[E]) Log(l *slog.Logger) *EventError[E] {
	dh := &defaultErrHandler{h}
	dh.Log(l)
	return h
}

func (h *EventError[E]) Reply(s *dgo.Session, r *dgo.Message) *EventError[E] {
	dh := &defaultErrHandler{h}
	dh.Reply(s, r)
	return h
}

func (h *EventError[E]) Send(s *dgo.Session, channelID string) *EventError[E] {
	dh := &defaultErrHandler{h}
	dh.Send(s, channelID)
	return h
}
