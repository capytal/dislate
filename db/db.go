package db

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Queries struct {
	db  DBTX
	ctx context.Context
}

func New(db DBTX, ctx ...context.Context) *Queries {
	var c context.Context
	if len(ctx) > 0 {
		c = ctx[0]
	} else {
		c = context.Background()
	}

	return &Queries{db, c}
}

func Prepare(db DBTX, ctx ...context.Context) (*Queries, error) {
	q := New(db, ctx...)

	if _, err := q.exec(guildCreate); err != nil {
		return nil, err
	}
	if _, err := q.exec(channelCreate); err != nil {
		return nil, err
	}
	if _, err := q.exec(messageCreate); err != nil {
		return nil, err
	}

	return q, nil
}

func (q *Queries) WithTx(tx *sql.Tx, ctx ...context.Context) *Queries {
	return New(tx, ctx...)
}

func (q *Queries) WithContext(ctx context.Context) *Queries {
	return New(q.db, ctx)
}

func (q *Queries) exec(query string, args ...any) (sql.Result, error) {
	return q.db.ExecContext(q.ctx, query, args...)
}

func (q *Queries) query(query string, args ...any) (*sql.Rows, error) {
	return q.db.QueryContext(q.ctx, query, args...)
}

func (q *Queries) queryRow(query string, args ...any) *sql.Row {
	return q.db.QueryRowContext(q.ctx, query, args...)
}
