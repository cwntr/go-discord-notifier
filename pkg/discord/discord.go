package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cwntr/go-discord-notifier/pkg/common"
)

var session *discordgo.Session

var (
	token string

	cfg common.Config

	newThreadChannelId    string
	updateThreadChannelId string

	newThreadCache []int
)

func InitDiscordSession(c common.Config) {
	var err error
	cfg = c
	token = c.DiscordToken
	session, err = discordgo.New("Bot " + c.DiscordToken)
	if err != nil {
		fmt.Printf("err while creating discord bot, err: %v", err)
		return
	}

	err = session.Open()
	if err != nil {
		fmt.Printf("err while creating discord bot, err: %v", err)
		return
	}
	newThreadChannelId = c.DiscordNewThreadChannelID
	updateThreadChannelId = c.DiscordUpdateThreadChannelID
}

func NotifyNewThread(id int, link string, subject string, content string) error {
	if session == nil {
		return fmt.Errorf("not initalizied")
	}
	for _, cid := range newThreadCache {
		if cid == id {
			return fmt.Errorf("notification already sent (can occur on error)")
		}
	}

	m := &discordgo.MessageEmbed{
		URL:   link,
		Title: fmt.Sprintf("/biz/ new thread [%d] %s", id, subject),
		Color: 0x0099ff,
	}
	if subject != "" && content != "" {
		m.Fields = []*discordgo.MessageEmbedField{{Name: subject, Value: content}}
	}
	_, err := session.ChannelMessageSendEmbed(newThreadChannelId, m)
	if err != nil {
		fmt.Printf("err:%v", err)
	} else {
		newThreadCache = append(newThreadCache, id)
	}
	return nil
}

func NotifyUpdateThread(id int, link string, subject string, replies int, replyAuthor string, replyContent string) error {
	if session == nil {
		return fmt.Errorf("not initalizied")
	}
	m := &discordgo.MessageEmbed{
		URL:         link,
		Title:       fmt.Sprintf("/biz/ update: [%d] %s", id, subject),
		Color:       0xffd433,
		Description: fmt.Sprintf("replies: [%d]", replies),
	}

	hasKeyword := false
	for _, k := range cfg.Keywords {
		if strings.Contains(strings.ToLower(replyContent), strings.ToLower(k)) {
			hasKeyword = true
		}
	}

	// reply must contain keyword
	if !hasKeyword {
		return nil
	}

	if replyAuthor != "" && replyContent != "" {
		m.Fields = []*discordgo.MessageEmbedField{{Name: fmt.Sprintf("[%s]", replyAuthor), Value: replyContent}}
	}
	_, err := session.ChannelMessageSendEmbed(updateThreadChannelId, m)
	if err != nil {
		fmt.Printf("err:%v", err)
	}
	return nil
}
