package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/tursodatabase/go-libsql"
)

const DEST_CHANNEL = "1270407366617333920"
const INPUT_CHANNEL = "1270407349869482006"
const USER_WEBHOOK_FORMAT = "dislate-user-%s"

func main() {
	log.Printf("Hello, world")

	db, err := sql.Open("libsql", "file:./guild.db")
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS MessageMap (
			Original text NOT NULL PRIMARY KEY UNIQUE,
			Translated text NOT NULL UNIQUE
		);
	`)
	if err != nil {
		log.Printf("ERROR: failed to create MessageMap table %s", err)
		return
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Printf("ERROR: failed to start bot %s", err)
		return
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

		wm, err := s.WebhookExecute(w.ID, w.Token, true, &discordgo.WebhookParams{
			Username:  m.Author.GlobalName,
			AvatarURL: m.Author.AvatarURL(""),
			Content:   m.Content,
		})
		if err != nil {
			log.Printf("ERROR: failed to message using webhook for user %s: %s", m.Author.ID, err)
			return
		}

		_, err = db.Exec(`
			INSERT INTO MessageMap (Original, Translated)
				VALUES ($1, $2) ON CONFLICT DO NOTHING
		`, m.ID, wm.ID)
		if err != nil {
			log.Printf("ERROR: failed add message to database. Original: %s, Translated: %s: %s", m.ID, wm.ID, err)
			return
		}
	})

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author.Bot {
			return
		}
		q := db.QueryRow(`
			SELECT * FROM MessageMap
				WHERE "Original" = $1
		`, m.ID)
		var original, translated string
		err := q.Scan(&original, &translated)
		if err != nil {
			log.Printf("ERROR: failed query message to database. Original: %s: %s", m.ID, err)
			return
		}

		ws, err := s.ChannelWebhooks(DEST_CHANNEL)
		if err != nil {
			log.Printf("ERROR: failed to find channel webhooks: %s", err)
			return
		}
		var w *discordgo.Webhook
		if wi := slices.IndexFunc(ws, func(w *discordgo.Webhook) bool {
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

		_, err = s.WebhookMessageEdit(w.ID, w.Token, translated, &discordgo.WebhookEdit{
			Content: &m.Content,
		})
		if err != nil {
			log.Printf("ERROR: failed to edit webhook message, Original %s, Translated %s: %s", m.ID, translated, err)
			return
		}
	})

	err = discord.Open()
	if err != nil {
		log.Printf("could not open session: %s", err)
		return
	}
	log.Printf("Bot session opened successfully")
	defer func() {
		err = discord.Close()
		if err != nil {
			log.Printf("could not close session: %s", err)
			return
		}
		log.Printf("Bot session closed successfully")
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT)
	<-sig
}
