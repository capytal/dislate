package gconf

import (
	"log/slog"

	gdb "dislate/internals/guilddb"

	dgo "github.com/bwmarrin/discordgo"
)

type Config struct {
	Logger *slog.Logger
}

type ConfigString struct {
	LoggingChannel *string     `json:"logging_channel"`
	LoggingLevel   *slog.Level `json:"logging_level"`
}

type (
	Guild gdb.Guild[ConfigString]
	DB    gdb.GuildDB[ConfigString]
)

func (g Guild) GetConfig(s *dgo.Session) (*Config, error) {
	var l *slog.Logger
	var err error

	if g.Config.LoggingChannel != nil {
		c, err := s.Channel(*g.Config.LoggingChannel)
		if err != nil {
			return nil, err
		}

		var lv slog.Level
		if g.Config.LoggingLevel != nil {
			lv = *g.Config.LoggingLevel
		} else {
			lv = slog.LevelInfo
		}
		l = slog.New(NewGuildHandler(s, c, &slog.HandlerOptions{
			Level: lv,
		}))
	} else {
		l = slog.New(disabledHandler{})
	}

	return &Config{l}, err
}

func GetLogger(guildID string, s *dgo.Session, db DB) *slog.Logger {
	g, err := db.Guild(guildID)
	if err != nil {
		return slog.New(disabledHandler{})
	}

	c, err := Guild(g).GetConfig(s)
	if err != nil {
		return slog.New(disabledHandler{})
	}

	return c.Logger
}
