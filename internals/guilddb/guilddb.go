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

type GuildDB interface {
	// Selects and returns a Message from the database, based on the
	// key pair of Channel's ID and Message's ID.
	//
	// Will return ErrNotFound if no message is found or ErrInternal.
	Message(channelID, messageID string) (Message, error)
	// Returns a slice of Messages with the provided Message.OriginChannelID and Message.OriginID.
	//
	// Will return ErrNotFound if no message is found (slice's length == 0) or ErrInternal.
	MessagesWithOrigin(originChannelID, originID string) ([]Message, error)
	// Returns a Messages with the provided Message.OriginChannelID, Message.OriginID
	// and Message.Language.
	//
	// Will return ErrNotFound if no message is found or ErrInternal.
	MessagesWithOriginByLang(originChannelId, originId string, language lang.Language) (Message, error)
	// Inserts a new Message object in the database.
	//
	// Message.ChannelID and Message.ID must be a unique pair and not already
	// in the database. Implementations of this function may require Message.ChannelID
	// to be an already stored Channel object, in the case that it isn't stored,
	// ErrPreconditionFailed may be returned.
	//
	// Message.OriginID and Message.OriginChannelID should not be nil if the message
	// is a translated one.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	MessageInsert(message Message) error
	// Updates the Message object in the database. Message.ID and Message.ChannelID
	// are used to find the correct message.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	MessageUpdate(message Message) error
	// Deletes the Message object in the database. Message.ID and Message.ChannelID
	// are used to find the correct message.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	MessageDelete(message Message) error
	// Selects and returns a Channel from the database, based on the
	// ID provided.
	//
	// Will return ErrNotFound if no channel is found or ErrInternal.
	Channel(channelID string) (Message, error)
	// Inserts a new Channel object in the database.
	//
	// Channel.ID must be unique and not already in the database.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	ChannelInsert(channel Channel) error
	// Updates the Channel object in the database. Channel.ID is used to find the
	// correct Channel.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelUpdate(channel Channel) error
	// Deletes the Channel object in the database. Channel.ID is used to find the
	// correct Channel.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelDelete(channel Channel) error
	// Selects and returns a ChannelGroup from the database. Finds a ChannelGroup
	// that has a Channel if the provided ID.
	//
	// Channels cannot be in two ChannelGroup at the same time.
	//
	// Will return ErrNotFound if no channel is found or ErrInternal.
	ChannelGroup(channelID string) (Message, error)
	// Inserts a new ChannelGroup object in the database. ChannelGroup must be unique
	// and not have Channels that are already in other groups.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	ChannelGroupInsert(group ChannelGroup) error
	// Updates the ChannelGroup object in the database.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelGroupUpdate(channel Channel) error
	// Deletes the ChannelGroup object in the database.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelGroupDelete(channel Channel) error
}

var ErrNoAffect = errors.New("Not able to affect anything in the database")
var ErrNotFound = errors.New("Object not found in the database")
var ErrPreconditionFailed = errors.New("Precondition failed")
var ErrInvalidObject = errors.New("Invalid object")
var ErrInternal = errors.New("Internal error while trying to use database")
