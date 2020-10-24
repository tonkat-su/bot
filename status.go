package main

import (
	"context"
	"fmt"
	"log"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/imgur"
	"github.com/tonkat-su/bot/mcuser"
	"github.com/vincent-petithory/dataurl"
)

func sendServerStatus(s *discordgo.Session, m *discordgo.MessageCreate, cfg Config, imgurClient *imgur.Client) error {
	ctx := context.Background()
	hostports, err := resolveMinecraftHostPort(ctx, nil, cfg.ServerHost)
	if err != nil {
		return fmt.Errorf("error resolving server host '%s': %s", cfg.ServerHost, err.Error())
	}
	if len(hostports) == 0 {
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title: cfg.ServerName,
	}

	serverUrl := hostports[0].String()
	pong, err := mcpinger.New(hostports[0].Host, hostports[0].Port).Ping()
	if err != nil {
		log.Printf("error pinging server '%s': %s", serverUrl, err.Error())
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "error",
				Value: err.Error(),
			},
		}
		embed.Color = 0xf04747
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "gamers?",
			Embed:   embed,
		})
		if err != nil {
			log.Printf("error sending message: %s", err.Error())
		}
		return nil
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "host",
			Value: serverUrl,
		},
	}

	players := make([]*Player, len(pong.Players.Sample))
	for i, p := range pong.Players.Sample {
		uuid, err := mcuser.GetUuid(p.Name)
		if err != nil {
			uuid = p.ID
			log.Printf("error getting uuid for user %s: %s", p.Name, err.Error())
		}
		players[i] = &Player{
			Name: p.Name,
			Uuid: uuid,
		}
	}
	playersEmbedField := &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("online (%d/%d)", pong.Players.Online, pong.Players.Max),
	}
	if len(players) == 0 {
		playersEmbedField.Value = ":("
	} else {
		err = syncMinecraftAvatarsToEmoji(s, m.GuildID, players)
		if err != nil {
			log.Printf("error syncing emoji: %s", err.Error())
		}

		playersEmbedField.Value = playerListEmojis(players)
	}
	embed.Fields = append(embed.Fields, playersEmbedField)
	embed.Color = 0x43b581

	embed.Description = pong.Description.Text

	if pong.Favicon != "" && imgurClient != nil {
		favIcon, err := dataurl.DecodeString(pong.Favicon)
		if err != nil {
			return fmt.Errorf("error decoding favicon for server '%s': %s", serverUrl, err.Error())
		}

		uploadRequest := &imgur.ImageUploadRequest{
			Image: favIcon.Data,
			Name:  serverUrl,
		}
		img, err := imgurClient.Upload(ctx, uploadRequest)
		if err != nil {
			return fmt.Errorf("error uploading favicon for server '%s' to imgur: %s", serverUrl, err.Error())
		}
		embed.Image = &discordgo.MessageEmbedImage{
			URL: img.Link,
		}
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "gamers?",
		Embed:   embed,
	})
	if err != nil {
		log.Printf("error sending message: %s", err.Error())
	}

	return nil
}

/*
func listServers(cfg Config, imgurClient *imgur.Client) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content != "list servers" {
			return
		}

		embeds := make([]*discordgo.MessageEmbed, 0, len(cfg.Servers))
		for serverName, host := range cfg.Servers {
			embed, err := prepareEmbedWithServerStatus(cfg, imgurClient, serverName, host)
			if err != nil {
				log.Printf("error preparing embed for server '%s': %s", serverName, err.Error())
				return
			}
			embeds = append(embeds, embed)
		}

		webhookParams := &discordgo.WebhookParams{
			Embeds: embeds,
		}
		_, err := s.WebhookExecute(cfg.DiscordWebhookId, cfg.DiscordWebhookToken, false, webhookParams)
		if err != nil {
			log.Printf("failed to execute webhook: %s", err.Error())
		}
	}
}
*/
