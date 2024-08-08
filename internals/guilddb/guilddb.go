package guilddb

import (
	"dislate/internals/lang"
	"errors"
)

type Message struct {
	ID              string
	ChannelID       string
	Language        lang.Language
	OriginID        *string
	OriginChannelID *string
}

type Channel struct {
	ID       string
	Language lang.Language
}

type ChannelGroup []Channel

type UserWebhook struct {
	ID        string
	ChannelID string
	UserID    string
	Token     string
}

type GuildDB interface {
}

var ErrNoAffect = errors.New("Not able to affect anything in the database")
var ErrNoMessages = errors.New("Messages not found in database")
var ErrNoChannels = errors.New("Channels not found in database")
var ErrNoChannelGroup = errors.New("Channel group not found in database")
var ErrInvalidChannelGroup = errors.New("Invalid Channel group in database (not sorted)")
var ErrMissingChannels = errors.New("Missing channels in database")
var ErrInternal = errors.New("Internal error while trying to use database")
