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
var ErrNotFound = errors.New("Object not found in the database")
var ErrPreconditionFailed = errors.New("Precondition failed")
var ErrInvalidObject = errors.New("Invalid object")
var ErrInternal = errors.New("Internal error while trying to use database")
