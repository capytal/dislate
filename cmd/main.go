package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forge.capytal.company/capytal/dislate"
	"forge.capytal.company/capytal/dislate/translator"

	_ "github.com/tursodatabase/go-libsql"

	"github.com/charmbracelet/log"
)

type TranslationProvider string

const (
	GOOGLE_TRANSLATE TranslationProvider = "google-translate"
)

// var translation_provider = flag.String("tprovider", string(GOOGLE_TRANSLATE), "Translation provider")
var (
	database_file = flag.String("db", "file:./guild.db", "SQLite database file/location")
	discord_token = flag.String(
		"token",
		os.Getenv("DISCORD_TOKEN"),
		"Discord bot authentication token",
	)
	verbose = flag.Bool("v", false, "Enable debug information.")
)

func init() {
	flag.Parse()
}

func main() {
	var logLevel log.Level
	if *verbose {
		logLevel = log.DebugLevel
	} else {
		logLevel = log.InfoLevel
	}
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{
		TimeFormat:      time.DateTime,
		Level:           logLevel,
		ReportTimestamp: true,
		ReportCaller:    true,
	}))

	db, err := sql.Open("libsql", *database_file)
	if err != nil {
		logger.Error("Failed to start SQLite database", slog.String("error", err.Error()))
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

	err = dislate.Run(*discord_token, dislate.RunOptions{
		DB:         db,
		Translator: translator.NewMockTranslator(),
		Logger:     logger,
	})
	if err != nil {
		logger.Error("Failed to start discord bot", slog.String("err", err.Error()))
		return
	}

	logger.Info("Discord bot started")

	defer func() {
		if err := dislate.Stop(); err != nil {
			logger.Error("Failed to stop discord bot", slog.String("err", err.Error()))
			return
		}
		logger.Info("Discord bot stopped")
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig
}
