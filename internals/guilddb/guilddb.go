package guilddb

import (
	"dislate/internals/translator/lang"
	"errors"
)

type Guild struct {
	ID string
}

type Channel struct {
	ID       string
	GuildID  string
	Language lang.Language
}
type ChannelGroup []Channel

type Message struct {
	ID              string
	ChannelID       string
	GuildID         string
	Language        lang.Language
	OriginID        *string
	OriginChannelID *string
}

type GuildDB interface {
	// Selects and returns a Message from the database, based on the
	// key pair of Channel's ID and Message's ID.
	//
	// Will return ErrNotFound if no message is found or ErrInternal.
	Message(guildID, channelID, ID string) (Message, error)
	// Returns a slice of Messages with the provided Message.OriginChannelID and Message.OriginID.
	//
	// Will return ErrNotFound if no message is found (slice's length == 0) or ErrInternal.
	MessagesWithOrigin(guildID, originChannelID, originID string) ([]Message, error)
	// Returns a Messages with the provided Message.OriginChannelID, Message.OriginID
	// and Message.Language.
	//
	// Will return ErrNotFound if no message is found or ErrInternal.
	MessageWithOriginByLang(guildID, originChannelId, originId string, language lang.Language) (Message, error)
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
	MessageInsert(m Message) error
	// Updates the Message object in the database. Message.ID and Message.ChannelID
	// are used to find the correct message.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	MessageUpdate(m Message) error
	// Deletes the Message object in the database. Message.ID and Message.ChannelID
	// are used to find the correct message.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	MessageDelete(m Message) error
	// Selects and returns a Channel from the database, based on the
	// ID provided.
	//
	// Will return ErrNotFound if no channel is found or ErrInternal.
	Channel(guildID, ID string) (Channel, error)
	// Inserts a new Channel object in the database.
	//
	// Channel.ID must be unique and not already in the database.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	ChannelInsert(c Channel) error
	// Updates the Channel object in the database. Channel.ID is used to find the
	// correct Channel.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelUpdate(c Channel) error
	// Deletes the Channel object in the database. Channel.ID is used to find the
	// correct Channel.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	ChannelDelete(c Channel) error
	// Selects and returns a ChannelGroup from the database. Finds a ChannelGroup
	// that has a Channel if the provided ID.
	//
	// Channels cannot be in two ChannelGroup at the same time.
	//
	// Will return ErrNotFound if no channel is found or ErrInternal.
	ChannelGroup(guildID, ID string) (ChannelGroup, error)
	// Inserts a new ChannelGroup object in the database. ChannelGroup must be unique
	// and not have Channels that are already in other groups.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	ChannelGroupInsert(g ChannelGroup) error
	// Updates the ChannelGroup object in the database.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	ChannelGroupUpdate(g ChannelGroup) error
	// Deletes the ChannelGroup object in the database.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	ChannelGroupDelete(g ChannelGroup) error
	// Selects and returns a Guild from the database.
	//
	// Will return ErrNotFound if no Guild is found or ErrInternal.
	Guild(ID string) (Guild, error)
	// Inserts a new Guild object in the database. Guild.ID must be unique and
	// not already in the database.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	GuildInsert(g Guild) error
	// Delete a Guild from the database. Guild.ID is used to find the object.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	GuildDelete(g Guild) error
}

var ErrNoAffect = errors.New("Not able to affect anything in the database")
var ErrNotFound = errors.New("Object not found in the database")
var ErrPreconditionFailed = errors.New("Precondition failed")
var ErrInvalidObject = errors.New("Invalid object")
var ErrInternal = errors.New("Internal error while trying to use database")
