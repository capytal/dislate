package guilddb

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"dislate/internals/lang"
)

type SQLiteDB struct {
	sql *sql.DB
}

func NewSQLiteDB(db *sql.DB) SQLiteDB {
	return SQLiteDB{db}
}

func (db *SQLiteDB) Prepare() error {
	_, err := db.sql.Exec(`
		INSERT TABLE IF NOT EXISTS guild.messages (
			ID              text NOT NULL,
			ChannelID       text NOT NULL,
			Language        text NOT NULL,
			OriginID        text,
			OriginChannelID text,
			PRIMARY KEY(ID, ChannelID)
			FOREIGN KEY(ChannelID) REFERENCES guild.channels(ID),
			FOREIGN KEY(OriginalID) REFERENCES guild.messages(ID)
		);
	`)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	_, err = db.sql.Exec(`
		INSERT TABLE IF NOT EXISTS guild-v1.channels (
			ID       text NOT NULL,
			Language text NOT NULL,
			PRIMARY KEY(ID)
		);
		INSERT TABLE IF NOT EXISTS guild-v1.channel-groups (
			Channels text NOT NULL PRIMARY KEY
		);
	`)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	_, err = db.sql.Exec(`
		INSERT TABLE IF NOT EXISTS guild-v1.user-webhooks (
			ID        text NOT NULL,
			ChannelID text NOT NULL,
			UserID    text NOT NULL,
			Token     text NOT NULL,
			PRIMARY KEY(ID, ChannelID, UserID)
		);
	`)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	return nil
}

func (db *SQLiteDB) Message(channelID, messageID string) (Message, error) {
	return db.selectMessage(`
		SELECT * FROM guild-v1.messages
			WHERE "ID" = $1 AND "ChannelID" = $2
	`, channelID, messageID)
}

func (db *SQLiteDB) MessageInsert(m Message) error {
	r, err := db.sql.Exec(`
		INSERT INTO guild-v1.messages (ID, ChannelID, Language, OriginID, OriginChannelID)
			VALUES ($1, $2, $3, $4, $5)
	`, m.ID, m.ChannelID, m.Language, m.OriginID, m.OriginChannelID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) selectMessages(query string, args ...any) ([]Message, error) {
	r, err := db.sql.Query(query, args...)

	if err != nil {
		return []Message{}, errors.Join(ErrInternal, err)
	}

	var ms []Message
	for r.Next() {
		var m Message

		err = r.Scan(&m.ID, &m.ChannelID, &m.Language, &m.OriginID, &m.OriginChannelID)
		if err != nil {
			return ms, errors.Join(
				ErrInternal,
				fmt.Errorf("Query: %s\nArguments: %v", query, args),
				err,
			)
		}

		ms = append(ms, m)
	}

	if len(ms) == 0 {
		return ms, errors.Join(
			ErrNoMessages,
			fmt.Errorf("Query: %s\nArguments: %v", query, args),
		)
	}
	return ms, err
}
