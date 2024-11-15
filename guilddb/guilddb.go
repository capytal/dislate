package guilddb

import (
	"errors"

	"forge.capytal.company/capytal/dislate/translator"
)

type Guild[C any] struct {
	ID     string
	Config C
}

func NewGuild[C any](ID string, config C) Guild[C] {
	return Guild[C]{ID, config}
}

type Channel struct {
	GuildID  string
	ID       string
	Language translator.Language
}

func NewChannel(GuildID, ID string, lang translator.Language) Channel {
	return Channel{GuildID, ID, lang}
}

type ChannelGroup []Channel

type Message struct {
	GuildID         string
	ChannelID       string
	ID              string
	Language        translator.Language
	OriginChannelID *string
	OriginID        *string
}

func NewMessage(GuildID, ChannelID, ID string, lang translator.Language) Message {
	return Message{GuildID, ChannelID, ID, lang, nil, nil}
}

func NewTranslatedMessage(
	GuildID, ChannelID, ID string,
	lang translator.Language,
	OriginChannelID, OriginID string,
) Message {
	return Message{GuildID, ChannelID, ID, lang, &OriginChannelID, &OriginID}
}

type GuildDB[C any] interface {
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
	MessageWithOriginByLang(
		guildID, originChannelId, originId string,
		language translator.Language,
	) (Message, error)
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
	// Deletes all messages in a Channel in the database. Channel.ID is used to find
	// the correct messages.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	MessageDeleteFromChannel(c Channel) error
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
	Guild(ID string) (Guild[C], error)
	// Inserts a new Guild object in the database. Guild.ID must be unique and
	// not already in the database.
	//
	// Will return ErrNoAffect if the object already exists or ErrInternal.
	GuildInsert(g Guild[C]) error
	// Delete a Guild from the database. Guild.ID is used to find the object.
	//
	// Will return ErrNoAffect if no object was deleted or ErrInternal.
	GuildDelete(g Guild[C]) error
	// Updates the Guild object in the database.
	//
	// Will return ErrNoAffect if no object was updated or ErrInternal.
	GuildUpdate(g Guild[C]) error
}

var (
	ErrNoAffect           = errors.New("Not able to affect anything in the database")
	ErrNotFound           = errors.New("Object not found in the database")
	ErrPreconditionFailed = errors.New("Precondition failed")
	ErrInvalidObject      = errors.New("Invalid object")
	ErrInternal           = errors.New("Internal error while trying to use database")
	ErrConfigParsing      = errors.New("Error while parsing Guild's config")
)
