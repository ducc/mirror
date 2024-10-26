package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	token           = os.Getenv("TOKEN")
	sourceChannelId = os.Getenv("CHANNEL_ID")
	webhookUrl      = os.Getenv("WEBHOOK_URL")
)

func main() {
	discord, err := discordgo.New(token)
	if err != nil {
		panic(err)
	}

	discord.UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0"

	discord.AddHandler(messageCreate)

	if err := discord.Open(); err != nil {
		panic(err)
	}

	fmt.Println("k")

	c := make(chan struct{})
	<-c
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != sourceChannelId {
		return
	}

	params := discordgo.WebhookParams{
		Content:         m.Message.Content,
		Components:      m.Message.Components,
		Username:        m.Author.Username,
		AvatarURL:       m.Author.AvatarURL(""),
		Embeds:          m.Message.Embeds,
		Attachments:     m.Message.Attachments,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}

	data, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	res, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(respBody))
	}
}
