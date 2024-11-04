package db

type Guild struct {
	ID string
}

const guildCreate = `
CREATE IF NOT EXISTS guilds (
	ID     text NOT NULL,
	PRIMARY KEY(ID)
);
`

const createGuild = `
INSERT INTO guilds (ID) VALUES (?);
`

func (q *Queries) CreateGuild(id string) error {
	_, err := q.exec(createGuild, id)
	return err
}

const updateGuild = `
UPDATE guilds SET ID = ? WHERE ID = ?;
`

func (q *Queries) UpdateGuild(id string) error {
	_, err := q.exec(updateGuild, id, id)
	return err
}

const getGuild = `
SELECT (ID) FROM guilds WHERE ID = ?;
`

func (q *Queries) GetGuild(id string) (Guild, error) {
	row := q.queryRow(getGuild, id)

	var g Guild
	if err := row.Scan(&g.ID); err != nil {
		return Guild{}, err
	}

	return g, nil
}
