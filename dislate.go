package dislate

import (
	"database/sql"
	"log/slog"

	bot "forge.capytal.company/capytal/dislate/lib"
	"forge.capytal.company/capytal/dislate/translator"
)

var dislate *bot.Bot

type RunOptions struct {
	DB         *sql.DB
	Translator translator.Translator
	Logger     *slog.Logger
}

func Start(token string, opts ...RunOptions) error {
	var err error

	dislate, err = bot.Start(token)
	if err != nil {
		return err
	}

	return nil
}

func Stop() error { return dislate.Stop() }
