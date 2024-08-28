package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dislate/internals/discord/bot"
	"dislate/internals/discord/bot/gconf"
	"dislate/internals/guilddb"
	"dislate/internals/translator"

	"github.com/charmbracelet/log"
)

type TranslationProvider string

const (
	GOOGLE_TRANSLATE TranslationProvider = "google-translate"
)

// var translation_provider = flag.String("tprovider", string(GOOGLE_TRANSLATE), "Translation provider")
var (
	database_file = flag.String("db", "file:./guild.db", "SQLite database file/location")
	discord_token = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Discord bot authentication token")
)

func init() {
	flag.Parse()
}

func main() {
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{
		TimeFormat:      time.DateTime,
		ReportTimestamp: true,
		ReportCaller:    true,
	}))

	db, err := guilddb.NewSQLiteDB[gconf.ConfigString](*database_file)
	if err != nil {
		logger.Error("Failed to open database connection", slog.String("err", err.Error()))
		return
	}
	logger.Info("Connection to database started", slog.String("file", *database_file))
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close database connection", slog.String("err", err.Error()))
			return
		}
		logger.Info("Connection to database closed", slog.String("file", *database_file))
	}()

	if err := db.Prepare(); err != nil {
		logger.Error("Failed to prepare database", slog.String("err", err.Error()))
		return
	}
	logger.Info("Database ready to be used")

	bot, err := bot.NewBot(*discord_token, db, translator.NewMockTranslator(), logger)
	if err != nil {
		logger.Error("Failed to create discord bot", slog.String("err", err.Error()))
		return
	}
	if err := bot.Start(); err != nil {
		logger.Error("Failed to start discord bot", slog.String("err", err.Error()))
		return
	}
	logger.Info("Discord bot started")
	defer func() {
		if err := bot.Stop(); err != nil {
			logger.Error("Failed to stop discord bot", slog.String("err", err.Error()))
			return
		}
		logger.Info("Discord bot stopped")
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig
}
