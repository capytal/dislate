package guilddb

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"dislate/internals/translator/lang"
	_ "github.com/tursodatabase/go-libsql"
)

type SQLiteDB struct {
	sql *sql.DB
}

func NewSQLiteDB(file string) (*SQLiteDB, error) {
	db, err := sql.Open("libsql", file)
	if err != nil {
		return &SQLiteDB{}, err
	}
	return &SQLiteDB{db}, nil
}

func (db *SQLiteDB) Close() error {
	return db.sql.Close()
}

func (db *SQLiteDB) Prepare() error {
	if _, err := db.sql.Exec(`
		CREATE TABLE IF NOT EXISTS guilds (
			ID text NOT NULL,
			PRIMARY KEY(ID)
		);
	`); err != nil {
		return errors.Join(ErrInternal, err)
	}

	if _, err := db.sql.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			GuildID  text NOT NULL,
			ID       text NOT NULL,
			Language text NOT NULL,
			PRIMARY KEY(ID, GuildID),
			FOREIGN KEY(GuildID) REFERENCES guilds(ID)
		);
		CREATE TABLE IF NOT EXISTS channelGroups (
			GuildID  text NOT NULL,
			Channels text NOT NULL,
			PRIMARY KEY(Channels, GuildID),
			FOREIGN KEY(GuildID) REFERENCES guilds(ID)
		);
	`); err != nil {
		return errors.Join(ErrInternal, err)
	}

	if _, err := db.sql.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			GuildID         text NOT NULL,
			ChannelID       text NOT NULL,
			ID              text NOT NULL,
			Language        text NOT NULL,
			OriginChannelID text,
			OriginID        text,
			PRIMARY KEY(ID, ChannelID, GuildID),
			FOREIGN KEY(ChannelID) REFERENCES channels(ID),
			FOREIGN KEY(GuildID) REFERENCES guilds(ID),
			FOREIGN KEY(OriginID) REFERENCES messages(ID),
			FOREIGN KEY(OriginChannelID) REFERENCES channels(ID)
		);
	`); err != nil {
		return errors.Join(ErrInternal, err)
	}

	return nil
}

func (db *SQLiteDB) Message(guildID, channelID, messageID string) (Message, error) {
	return db.selectMessage(`
		WHERE "GuildID" = $1 AND "ChannelID" = $2 AND "ID" = $3
	`, guildID, channelID, messageID)
}

func (db *SQLiteDB) MessagesWithOrigin(guildID, originChannelID, originID string) ([]Message, error) {
	return db.selectMessages(`
		WHERE "GuildID" = $1 AND "OriginChannelID" = $2 AND "OriginID" = $3
	`, guildID, originChannelID, originID)
}

func (db *SQLiteDB) MessageWithOriginByLang(
	guildID, originChannelID, originID string,
	language lang.Language,
) (Message, error) {
	return db.selectMessage(`
		WHERE "GuildID" = $1 AND "OriginChannelID" = $2 AND "OriginID" = $3  AND "Language" = $4
	`, guildID, originChannelID, originID, language)
}

func (db *SQLiteDB) MessageInsert(m Message) error {
	_, err := db.Channel(m.GuildID, m.ChannelID)
	if errors.Is(err, ErrNotFound) {
		return errors.Join(
			ErrPreconditionFailed,
			fmt.Errorf("Channel %s doesn't exists in the database", m.ChannelID),
		)
	} else if err != nil {
		return errors.Join(
			ErrInternal,
			errors.New("Failed to check if Channel exists in the database"),
			err,
		)
	}

	r, err := db.sql.Exec(`
		INSERT OR IGNORE INTO messages (GuildID, ChannelID, ID, Language, OriginChannelID, OriginID)
			VALUES ($1, $2, $3, $4, $5, $6)
	`, m.GuildID, m.ChannelID, m.ID, m.Language, m.OriginChannelID, m.OriginID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) MessageUpdate(m Message) error {
	r, err := db.sql.Exec(`
		UPDATE messages
			SET Language = $1, OriginChannelID = $2, OriginID = $3
			WHERE "GuildID" = $4 AND "ChannelID" = $5 AND "ID" = $6
	`, m.Language,
		m.OriginChannelID,
		m.OriginID,
		m.GuildID,
		m.ChannelID,
		m.ID,
	)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) MessageDelete(m Message) error {
	_, err := db.sql.Exec(`
		DELETE channels
			WHERE "GuildID" = $1 AND "OriginChannelID" = $2 AND "OriginID" = $3
	`, m.GuildID, m.ChannelID, m.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	r, err := db.sql.Exec(`
		DELETE channels
			WHERE "GuildID" = $1 AND "ChannelID" = $2 AND "ID" = $3
	`, m.GuildID, m.ChannelID, m.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) selectMessage(query string, args ...any) (Message, error) {
	var m Message
	err := db.sql.QueryRow(fmt.Sprintf(`
		SELECT GuildID, ChannelID, ID, Language, OriginChannelID, OriginID FROM messages
			%s
	`, query), args...).
		Scan(&m.GuildID, &m.ChannelID, &m.ID, &m.Language, &m.OriginChannelID, &m.OriginID)

	if errors.Is(err, sql.ErrNoRows) {
		return m, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return m, errors.Join(ErrInternal, err)
	}

	return m, nil
}

func (db *SQLiteDB) selectMessages(query string, args ...any) ([]Message, error) {
	r, err := db.sql.Query(fmt.Sprintf(`
		SELECT GuildID, ChannelID, ID, Language, OriginChannelID, OriginID FROM messages
			%s
	`, query), args...)

	if err != nil {
		return []Message{}, errors.Join(ErrInternal, err)
	}

	var ms []Message
	for r.Next() {
		var m Message

		err = r.Scan(&m.GuildID, &m.ChannelID, &m.ID, &m.Language, &m.OriginChannelID, &m.OriginID)
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

func (db *SQLiteDB) Channel(guildID, ID string) (Channel, error) {
	return db.selectChannel(`
		WHERE "GuildID" = $1 AND "ID" = $2
	`, guildID, ID)
}

func (db *SQLiteDB) ChannelInsert(c Channel) error {
	r, err := db.sql.Exec(`
		INSERT OR IGNORE INTO channels (GuildID, ID, Language)
			VALUES ($1, $2, $3)
	`, c.GuildID, c.ID, c.Language)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelUpdate(c Channel) error {
	r, err := db.sql.Exec(`
		UPDATE channels
			SET Language = $1
			WHERE "GuildID" = $2 AND "ID" = $3
	`, c.Language, c.GuildID, c.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelDelete(c Channel) error {
	r, err := db.sql.Exec(`
		DELETE channels
			WHERE "GuildID" = $1 AND "ID" = $2
	`, c.ID, c.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroup(guildID, channelID string) (ChannelGroup, error) {
	var g string

	err := db.sql.QueryRow(`
		SELECT GuildID, ID, Language FROM channelGroups
			WHERE "GuildID" = $1 AND "Channels" LIKE "%$2%"
	`, guildID, channelID).Scan(&g)

	if errors.Is(err, sql.ErrNoRows) {
		return ChannelGroup{}, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return ChannelGroup{}, errors.Join(ErrInternal, err)
	}

	ids := strings.Split(g, ",")
	if !slices.IsSorted(ids) {
		return ChannelGroup{}, errors.Join(
			ErrInvalidObject,
			fmt.Errorf("Channel in database is invalid, ids are not sorted: %s", ids),
		)
	}
	for i, v := range ids {
		ids[i] = fmt.Sprintf("\"ID\" = %s", v)
	}

	cs, err := db.selectChannels(fmt.Sprintf(`
		WHERE %s AND "GuildID" = $1
	`, strings.Join(ids, " OR ")), guildID)

	if errors.Is(err, ErrNotFound) || len(cs) != len(ids) {
		return ChannelGroup{}, errors.Join(
			ErrPreconditionFailed,
			fmt.Errorf("ChannelGroup has Channels that doesn't exist in the database, group: %s", ids),
			err,
		)
	} else if err != nil {
		return ChannelGroup{}, errors.Join(ErrInternal, err)
	}

	return cs, nil
}

func (db *SQLiteDB) ChannelGroupInsert(g ChannelGroup) error {
	if len(g) != 0 {
		return nil
	}

	var ids []string
	for _, c := range g {
		ids = append(ids, c.ID)
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(`
		INSERT OR IGNORE INTO channelGroups (GuildID, Channels)
			VALUES ($1, $2)
	`, g[0].GuildID, strings.Join(ids, ","))

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroupUpdate(g ChannelGroup) error {
	if len(g) != 0 {
		return nil
	}

	var ids, idsq []string
	for _, c := range g {
		ids = append(ids, c.ID)
		idsq = append(idsq, "\"ID\" LIKE \""+c.ID+"\"")
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(
		fmt.Sprintf(`
			UPDATE channelGroups
				SET Channels = $1
				WHERE %s AND "GuildID" = $2
		`, strings.Join(idsq, " OR ")),
		strings.Join(ids, ","),
		g[0].GuildID,
	)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) ChannelGroupDelete(g ChannelGroup) error {
	if len(g) != 0 {
		return nil
	}

	var ids, idsq []string
	for _, c := range g {
		ids = append(ids, c.ID)
		idsq = append(idsq, "\"ID\" LIKE \""+c.ID+"\"")
	}
	slices.Sort(ids)

	r, err := db.sql.Exec(
		fmt.Sprintf(`
			DELETE FROM channelGroups
				WHERE %s AND "GuildID" = $1
		`, strings.Join(idsq, " OR ")),
		g[0].GuildID,
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
	err := db.sql.QueryRow(fmt.Sprintf(`
		SELECT GuildID, ID, Language FROM channels
			%s
	`, query), args...).Scan(&c.GuildID, &c.ID, &c.Language)

	if errors.Is(err, sql.ErrNoRows) {
		return c, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return c, errors.Join(ErrInternal, err)
	}

	return c, nil
}

func (db *SQLiteDB) selectChannels(query string, args ...any) ([]Channel, error) {
	r, err := db.sql.Query(fmt.Sprintf(`
		SELECT GuildID, ID, Language FROM channels
			%s
	`, query), args...)

	if err != nil {
		return []Channel{}, errors.Join(ErrInternal, err)
	}

	var cs []Channel
	for r.Next() {
		var c Channel

		err = r.Scan(&c.GuildID, &c.ID, &c.Language)
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

func (db *SQLiteDB) Guild(ID string) (Guild, error) {
	var g Guild

	if err := db.sql.QueryRow(`
		SELECT "ID" FROM guilds
			WHERE "ID" = $1
	`, ID).Scan(g.ID); err != nil && errors.Is(err, sql.ErrNoRows) {
		return Guild{}, errors.Join(ErrNotFound, err)
	} else if err != nil {
		return Guild{}, errors.Join(ErrInternal, err)
	}

	return g, nil
}

func (db *SQLiteDB) GuildInsert(g Guild) error {
	r, err := db.sql.Exec(`
		INSERT OR IGNORE INTO guilds (ID)
			VALUES ($1)
	`, g.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}

func (db *SQLiteDB) GuildDelete(g Guild) error {
	r, err := db.sql.Exec(`
		DELETE FROM guilds
			WHERE "ID" = $1
	`, g.ID)

	if err != nil {
		return errors.Join(ErrInternal, err)
	} else if rows, _ := r.RowsAffected(); rows == 0 {
		return ErrNoAffect
	}

	return nil
}
