package errors

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
)

type defaultEventErr[E any] struct {
	message          string
	data             map[string]any
	session          *dgo.Session
	channelID        string
	messageReference *dgo.MessageReference
	logger           *slog.Logger
	errs             []error
}

func (d *defaultEventErr[E]) Join(errs ...error) EventErr {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &defaultEventErr[E]{
		message:          d.message,
		data:             d.data,
		session:          d.session,
		channelID:        d.channelID,
		messageReference: d.messageReference,
		logger:           d.logger,
		errs:             make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

func (d *defaultEventErr[E]) Error() string {
	var data []string
	for k, v := range d.data {
		data = append(data, slog.Any(k, v).String())
	}

	var e string
	if d.message != "" {
		e = fmt.Sprintf("%s-ERRO: %s %s", d.message, d.Event(), strings.Join(data, " "))
	} else {
		e = fmt.Sprintf("%s-ERRO: %s", d.Event(), strings.Join(data, " "))
	}

	var s strings.Builder
	_, berr := s.WriteString(e)
	if berr != nil {
		return "Failed to write error string"
	}
	for _, err := range d.errs {
		_, berr := s.WriteString("\n" + err.Error())
		if berr != nil {
			return "Failed to write error string"
		}
	}
	return s.String()
}

func (d *defaultEventErr[E]) Event() string {
	var e E
	t := reflect.TypeOf(e)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return strings.ToUpper(t.Name())
}

func (d *defaultEventErr[E]) Reply() error {
	if d.channelID == "" || d.messageReference == nil || d.session == nil {
		return nil
	}

	_, err := d.session.ChannelMessageSendReply(d.channelID, d.Error(), d.messageReference)
	return err
}

func (d *defaultEventErr[E]) Send() error {
	if d.channelID == "" || d.session == nil {
		return nil
	}

	_, err := d.session.ChannelMessageSend(d.channelID, d.Error())
	return err
}

func (d *defaultEventErr[E]) Log() {
	var args []any
	for k, v := range d.data {
		args = append(args, slog.Any(k, v))
	}

	d.logger.Error(d.Error(), args...)
}

func (d *defaultEventErr[E]) AddData(key string, v any) {
	d.data[key] = v
}
