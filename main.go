package main

import (
	"dislate/internals/discord"
	"dislate/internals/guilddb"
	"dislate/internals/translator"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const DEST_CHANNEL = "1270407366617333920"
const INPUT_CHANNEL = "1270407349869482006"
const USER_WEBHOOK_FORMAT = "dislate-user-%s"

func main() {
	log.Printf("Hello, world")

	db, err := guilddb.NewSQLiteDB("file:./guild.db")
	if err != nil {
		log.Printf("ERROR: failed to open database %s", err)
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("ERROR: failed to close database %s", err)
			return
		}
	}()

	if err := db.Prepare(); err != nil {
		log.Printf("ERROR: failed to prepare database: %s", err)
		return
	}

	bot, err := discord.NewBot(os.Getenv(""), db, translator.NewMockTranslator())
	if err != nil {
		log.Printf("ERROR: failed to create discord bot: %s", err)
		return
	}
	if err := bot.Start(); err != nil {
		log.Printf("ERROR: failed to start discord bot: %s", err)
		return
	}
	defer func() {
		if err := bot.Stop(); err != nil {
			log.Printf("ERROR: failed to stop discord bot: %s", err)
			return
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig
}
