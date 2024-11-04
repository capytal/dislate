package db

import (
	"encoding/json"

	"forge.capytal.company/capytal/dislate/translator"
)

type Message struct {
	GuildID        string
	ChannelID      string
	ID             string
	Language       translator.Language
	LinkedMessages []string
}

const messageCreate = `
CREATE IF NOT EXISTS messages (
	GuildID        text NOT NULL,
	ChannelID      text NOT NULL,
	ID             text NOT NULL,
	Language       text NOT NULL,
	LinkedMessages text NOT NULL,
	PRIMARY KEY(GuildID, ChannelID, ID),
	FOREIGN KEY(GuildID) REFERENCES guilds(ID),
	FOREIGN KEY(ChannelID) REFERENCES channels(ID)
);
`

const createMessage = `
INSERT INTO messages (
	GuildID,
	ChannelID,
	ID,
	Language,
	LinkedMessages,
) VALUES (?, ?, ?, ?, ?);
`

func (q *Queries) CreateMessage(
	GuildID, ChannelID, ID string,
	Language translator.Translator,
	LinkedMessages []string,
) error {
	j, err := json.Marshal(LinkedMessages)
	if err != nil {
		return err
	}

	_, err = q.exec(createMessage, GuildID, ChannelID, ID, Language, string(j))

	return err
}

const updateMessage = `
UPDATE messages
	SET GuildID = ?, ChannelID = ?, ID = ?, Language = ?, LinkedMessages = json(?)
	WHERE GuildID = ? AND ChannelID = ? AND ID = ?;
`

func (q *Queries) UpdateMessage(
	GuildID, ChannelID, ID string,
	Language translator.Translator,
	LinkedMessages []string,
) error {
	j, err := json.Marshal(LinkedMessages)
	if err != nil {
		return err
	}

	_, err = q.exec(
		updateMessage,
		GuildID,
		ChannelID,
		ID,
		Language,
		string(j),
		GuildID,
		ChannelID,
		ID,
	)

	return err
}

const getMessage = `
SELECT (GuildID, ChannelID, ID, Language, LinkedMessages) FROM messages
	WHERE GuildID = ? AND ChannelID = ? AND ID = ?;
`

func (q *Queries) GetMessage(GuildID, ChannelID, ID string) (Message, error) {
	row := q.queryRow(getMessage, GuildID, ChannelID, ID)

	var m Message
	var lm string

	if err := row.Scan(&m.GuildID, &m.ChannelID, &m.ID, &m.Language, &lm); err != nil {
		return m, err
	}

	if err := json.Unmarshal([]byte(lm), &m.LinkedMessages); err != nil {
		return m, err
	}

	return m, nil
}

const listChannelMessages = `
SELECT (GuildID, ChannelID, ID, Language, LinkedMessages) FROM messages
	WHERE GuildID = ? AND ChannelID = ?;
`

func (q *Queries) ListChannelMessages(GuildID, ChannelID string) ([]Message, error) {
	rows, err := q.query(listChannelMessages, GuildID, ChannelID)
	if err != nil {
		return []Message{}, err
	}

	ms := []Message{}

	for rows.Next() {
		var m Message
		var lm string

		if err := rows.Scan(&m.GuildID, &m.ChannelID, &m.ID, &m.Language, &lm); err != nil {
			return ms, err
		}

		if err := json.Unmarshal([]byte(lm), &m.LinkedMessages); err != nil {
			return ms, err
		}

		ms = append(ms, m)
	}

	return ms, err
}
