package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	log.Printf("Hello, world")

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}
	err = discord.Open()
	if err != nil {
		panic(err)
	}
}
