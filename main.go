package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Printf("Hello, world")

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID == "1270407349869482006" {
			_, err := s.ChannelMessageSend("1270407366617333920", m.Content)
			log.Print(m.Content)
			if err != nil {
				panic(err)
			}
		}
	})

	err = discord.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}
	log.Printf("Bot session opened successfully")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig

	err = discord.Close()
	if err != nil {
		log.Fatalf("could not close session: %s", err)
	}
	log.Printf("Bot session closed successfully")
}
