package db

import (
	"encoding/json"

	"forge.capytal.company/capytal/dislate/translator"
)

type Channel struct {
	GuildID        string
	ID             string
	Language       translator.Language
	LinkedChannels []string
}

const channelCreate = `
CREATE IF NOT EXISTS channels (
	GuildID        text NOT NULL,
	ID             text NOT NULL,
	Language       text NOT NULL,
	LinkedChannels text NOT NULL,
	PRIMARY KEY(GuildID, ID),
	FOREIGN KEY(GuildID) REFERENCES guilds(ID)
);
`

const createChannel = `
INSERT INTO channels (
	GuildID,
	ID,
	Language,
	LinkedChannels,
) VALUES (?, ?, ?, ?);
`

func (q *Queries) CreateChannel(
	GuildID, ID string,
	Language translator.Translator,
	LinkedChannels []string,
) error {
	j, err := json.Marshal(LinkedChannels)
	if err != nil {
		return err
	}

	_, err = q.exec(createChannel, GuildID, ID, Language, string(j))

	return err
}

const updateChannel = `
UPDATE channels
	SET GuildID = ?, ID = ?, Language = ?, LinkedChannels = json(?)
	WHERE GuildID = ? AND ID = ?;
`

func (q *Queries) UpdateChannel(
	GuildID, ID string,
	Language translator.Translator,
	LinkedChannels []string,
) error {
	j, err := json.Marshal(LinkedChannels)
	if err != nil {
		return err
	}

	_, err = q.exec(updateChannel, GuildID, ID, Language, string(j), GuildID, ID)

	return err
}

const getChannel = `
SELECT (GuildID, ID, Language, LinkedChannels) FROM channels
	WHERE GuildID = ? AND ID = ?;
`

func (q *Queries) GetChannel(GuildID, ID string) (Channel, error) {
	row := q.queryRow(getChannel, GuildID, ID)

	var c Channel
	var lc string

	if err := row.Scan(&c.GuildID, &c.ID, &c.Language, &lc); err != nil {
		return c, err
	}

	if err := json.Unmarshal([]byte(lc), &c.LinkedChannels); err != nil {
		return c, err
	}

	return c, nil
}

const listGuildChannels = `
SELECT (GuildID, ID, Language, LinkedChannels) FROM channels
	WHERE GuildID = ?;
`

func (q *Queries) ListGuildChannels(GuildID string) ([]Channel, error) {
	rows, err := q.query(listGuildChannels, GuildID)
	if err != nil {
		return []Channel{}, err
	}

	cs := []Channel{}

	for rows.Next() {
		var m Channel
		var lc string

		if err := rows.Scan(&m.GuildID, &m.ID, &m.Language, &lc); err != nil {
			return cs, err
		}

		if err := json.Unmarshal([]byte(lc), &m.LinkedChannels); err != nil {
			return cs, err
		}

		cs = append(cs, m)
	}

	return cs, err
}
