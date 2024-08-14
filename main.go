package main

import (
	"dislate/internals/discord"
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type TranslationProvider string

const (
	GOOGLE_TRANSLATE TranslationProvider = "google-translate"
)

var translation_provider = flag.String("tprovider", string(GOOGLE_TRANSLATE), "Translation provider")
var database_file = flag.String("db", "file:./guild.db", "SQLite database file/location")
var discord_token = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Discord bot authentication token")

func init() {
	flag.Parse()
}

func main() {
	log.Printf("Hello, world")

	db, err := guilddb.NewSQLiteDB(*database_file)
	if err != nil {
		log.Printf("ERROR: failed to open database %s", err)
		return
	}
	log.Print("Connection to database started")
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("ERROR: failed to close database %s", err)
			return
		}
		log.Print("Connection to database closed")
	}()

	if err := db.Prepare(); err != nil {
		log.Printf("ERROR: failed to prepare database: %s", err)
		return
	}
	log.Print("Database prepared to be used")

	bot, err := discord.NewBot(*discord_token, db, translator.NewMockTranslator())
	if err != nil {
		log.Printf("ERROR: failed to create discord bot: %s", err)
		return
	}
	if err := bot.Start(); err != nil {
		log.Printf("ERROR: failed to start discord bot: %s", err)
		return
	}
	log.Print("Connection to discord bot started")
	defer func() {
		if err := bot.Stop(); err != nil {
			log.Printf("ERROR: failed to stop discord bot: %s", err)
			return
		}
		log.Print("Connection to discord bot stopped")
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig
}
