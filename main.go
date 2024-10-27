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
	token      = os.Getenv("TOKEN")
	webhookUrl = os.Getenv("WEBHOOK_URL")

	// Guild to mirror messages from
	sourceGuildId = os.Getenv("SOURCE_GUILD_ID")
	// Channel to mirror messages from. Mirrors all channels if SOURCE_CHANNEL_ID is empty.
	sourceChannelId = os.Getenv("SOURCE_CHANNEL_ID")
	// Channel to mirror messages into
	targetChannelId = os.Getenv("TARGET_CHANNEL_ID")
)

var (
	// mapping from channel id to channel name
	channelNameCache = make(map[string]string)
)

func main() {
	if token == "" {
		panic("TOKEN is required")
	}
	if webhookUrl == "" {
		panic("WEBHOOK_URL is required")
	}
	if sourceGuildId == "" {
		panic("SOURCE_GUILD_ID is required")
	}
	if targetChannelId == "" {
		panic("TARGET_CHANNEL_ID is required")
	}

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

func getChannelName(s *discordgo.Session, channelID string) (string, error) {
	if cachedName := channelNameCache[channelID]; cachedName != "" {
		return cachedName, nil
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		return "", fmt.Errorf("getting channel %s: %w", channelID, err)
	}

	return channel.Name, nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages in target channel
	if m.ChannelID == targetChannelId {
		return
	}

	// ignore messages outside of source guild
	if sourceGuildId != m.GuildID {
		return
	}

	// if SOURCE_CHANNEL_ID is set then ignore messages from other channels
	if sourceChannelId != "" && m.ChannelID != sourceChannelId {
		return
	}

	channelName, err := getChannelName(s, m.ChannelID)
	if err != nil {
		panic(err)
	}

	params := discordgo.WebhookParams{
		Content:         m.Message.Content,
		Components:      m.Message.Components,
		Username:        truncate(fmt.Sprintf("%s - #%s", m.Author.Username, channelName), 80),
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

func truncate(inp string, maxLength int) string {
	if len(inp) <= maxLength {
		return inp
	}

	return inp[0:maxLength]
}
