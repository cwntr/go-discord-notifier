package common

import "time"

type Config struct {
	Keywords                     []string `json:"keywords"`
	DiscordToken                 string   `json:"discordToken"`
	DiscordNewThreadChannelID    string   `json:"discordNewThreadChannelID"`
	DiscordUpdateThreadChannelID string   `json:"discordUpdateThreadChannelID"`
	ProcessInterval              string   `json:"interval"`
	Interval                     time.Duration
}
