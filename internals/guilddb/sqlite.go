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
			FOREIGN KEY(OriginalID) REFERENCES guild.messages(ID),
			FOREIGN KEY(OriginalChannelID) REFERENCES guild.channels(ID)
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
	`, messageID, channelID)
}

func (db *SQLiteDB) MessagesWithOrigin(originID, originChannelID string) ([]Message, error) {
	return db.selectMessages(`
		SELECT * FROM guild-v1.messages
			WHERE "OriginID" = $1 AND "OriginChannelID" = $2
	`, originID, originChannelID)
}

func (db *SQLiteDB) MessageWithOriginByLang(originChannelID, originID string, language lang.Language) (Message, error) {
	return db.selectMessage(`
		SELECT * FROM guild-v1.messages
			WHERE "OriginID" = $1 AND "OriginChannelID" = $2 AND "Language" = $3
	`, originID, originChannelID, language)
}

func (db *SQLiteDB) MessageInsert(m Message) error {
	_, err := db.Channel(m.ChannelID)
	if errors.Is(err, ErrNotFound) {
		return errors.Join(ErrPreconditionFailed, fmt.Errorf("Channel %s doesn't exists in the database", m.ChannelID))
	} else if err != nil {
		return errors.Join(ErrInternal, errors.New("Failed to check if Channel exists in the database"), err)
	}

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

func (db *SQLiteDB) MessageUpdate(message Message) error {
	r, err := db.sql.Exec(`
		UPDATE guild-v1.messages
			SET Language = $1, OriginChannelID = $2, OriginID = $3
			WHERE "ID" = $4 AND "ChannelID" = $5
	`, message.Language,
		message.OriginChannelID,
		message.OriginID,
		message.ID,
		message.ChannelID,
	)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) MessageDelete(message Message) error {
	_, err := db.sql.Exec(`
		DELETE guild-v1.channels
			WHERE "OriginID" = $1 AND "OriginChannelID" = $2
	`, message.ID, message.ChannelID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	r, err := db.sql.Exec(`
		DELETE guild-v1.channels
			WHERE "ID" = $1 AND "ChannelID" = $2
	`, message.ID, message.ChannelID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) selectMessage(query string, args ...any) (Message, error) {
	var m Message
	err := db.sql.QueryRow(query, args...).
		Scan(&m.ID, &m.ChannelID, &m.Language, &m.OriginID, &m.OriginChannelID)

	if errors.Is(err, sql.ErrNoRows) {
		return m, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return m, errors.Join(ErrInternal, err)
	}

	return m, nil
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
			ErrNotFound,
			fmt.Errorf("Query: %s\nArguments: %v", query, args),
		)
	}
	return ms, err
}

func (db *SQLiteDB) Channel(channelID string) (Channel, error) {
	return db.selectChannel(`
		SELECT (ID, Language) FROM guild-v1.channels
			WHERE "ID" = $1
	`, channelID)
}

func (db *SQLiteDB) ChannelInsert(c Channel) error {
	r, err := db.sql.Exec(`
		INSERT INTO guild-v1.channels (ID, Language)
			VALUES ($1, $2)
	`, c.ID, c.Language)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelUpdate(channel Channel) error {
	r, err := db.sql.Exec(`
		UPDATE guild-v1.channels
			SET Language = $1
			WHERE "ID" = $2
	`, channel.Language, channel.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelDelete(channel Channel) error {
	r, err := db.sql.Exec(`
		DELETE guild-v1.channels
			WHERE "ID" = $1
	`, channel.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroup(channelID string) (ChannelGroup, error) {
	var g string

	err := db.sql.QueryRow(`
		SELECT (ID, Language) FROM guild-v1.channels
			WHERE "Channels" LIKE "%$1%"
	`, channelID).Scan(&g)

	if errors.Is(err, sql.ErrNoRows) {
		return ChannelGroup{}, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return ChannelGroup{}, errors.Join(ErrInternal, err)
	}

	ids := strings.Split(g, ",")
	if !slices.IsSorted(ids) {
		return ChannelGroup{}, ErrInvalidChannelGroup
	}
	for i, v := range ids {
		ids[i] = fmt.Sprintf("\"ID\" = %s", v)
	}

	cs, err := db.selectChannels(fmt.Sprintf(`
		SELECT (ID, Language) FROM guild-v1.channels
			WHERE %s
	`, strings.Join(ids, " OR ")))

	if errors.Is(err, ErrNotFound) || len(cs) != len(ids) {
		return ChannelGroup{}, errors.Join(ErrMissingChannels, err)
	} else if err != nil {
		return ChannelGroup{}, errors.Join(ErrInternal, err)
	}

	return cs, nil
}

func (db *SQLiteDB) ChannelGroupInsert(group ChannelGroup) error {
	var ids []string
	for _, c := range group {
		ids = append(ids, c.ID)
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(`
		INSERT INTO guild-v1.channel-groups (Channels)
			VALUES ($1)
	`, strings.Join(ids, ","))

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroupUpdate(group ChannelGroup) error {
	var ids, idsq []string
	for _, c := range group {
		ids = append(ids, c.ID)
		idsq = append(idsq, "\"ID\" LIKE \""+c.ID+"\"")
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(
		fmt.Sprintf(`
			UPDATE guild-v1.channel-groups
				SET Channels = $1
				WHERE %s
		`, strings.Join(idsq, " OR ")),
		strings.Join(ids, ","),
	)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroupDelete(group ChannelGroup) error {
	var ids, idsq []string
	for _, c := range group {
		ids = append(ids, c.ID)
		idsq = append(idsq, "\"ID\" LIKE \""+c.ID+"\"")
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(
		fmt.Sprintf(`
			DELETE FROM guild-v1.channel-groups
				WHERE %s
		`, strings.Join(idsq, " OR ")),
		strings.Join(ids, ","),
	)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) selectChannel(query string, args ...any) (Channel, error) {
	var c Channel
	err := db.sql.QueryRow(query, args...).
		Scan(&c.ID, &c.Language)

	if errors.Is(err, sql.ErrNoRows) {
		return c, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return c, errors.Join(ErrInternal, err)
	}

	return c, nil
}

func (db *SQLiteDB) selectChannels(query string, args ...any) ([]Channel, error) {
	r, err := db.sql.Query(query, args...)

	if err != nil {
		return []Channel{}, errors.Join(ErrInternal, err)
	}

	var cs []Channel
	for r.Next() {
		var c Channel

		err = r.Scan(&c.ID, &c.Language)
		if err != nil {
			return cs, errors.Join(
				ErrInternal,
				fmt.Errorf("Query: %s\nArguments: %v", query, args),
				err,
			)
		}

		cs = append(cs, c)
	}

	if len(cs) == 0 {
		return cs, errors.Join(
			ErrNotFound,
			fmt.Errorf("Query: %s\nArguments: %v", query, args),
		)
	}
	return cs, err
}
