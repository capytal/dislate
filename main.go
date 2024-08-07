package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const DEST_CHANNEL = "1270407366617333920"
const INPUT_CHANNEL = "1270407349869482006"
const USER_WEBHOOK_FORMAT = "dislate-user-%s"

func main() {
	log.Printf("Hello, world")

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID != INPUT_CHANNEL {
			return
		}
		if m.Author.Bot {
			return
		}

		ws, err := s.ChannelWebhooks(DEST_CHANNEL)
		if err != nil {
			log.Printf("ERROR: failed to find channel webhooks: %s", err)
			return
		}

		var w *discordgo.Webhook
		if wi := slices.IndexFunc(ws, func(w *discordgo.Webhook) bool {
			log.Printf(w.Name)
			return w.Name == fmt.Sprintf(USER_WEBHOOK_FORMAT, m.Author.ID)
		}); wi == -1 {
			w, err = s.WebhookCreate(
				DEST_CHANNEL,
				fmt.Sprintf(USER_WEBHOOK_FORMAT, m.Author.ID),
				m.Author.AvatarURL(""),
			)
			if err != nil {
				log.Printf("ERROR: failed to create webhook for user %s: %s", m.Author.ID, err)
				return
			}
		} else {
			w = ws[wi]
		}

		_, err = s.WebhookExecute(w.ID, w.Token, true, &discordgo.WebhookParams{
			Username: m.Author.GlobalName,
			Content:  m.Content,
		})
		if err != nil {
			log.Printf("ERROR: failed to message using webhook for user %s: %s", m.Author.ID, err)
			return
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
