package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/imgur"
	"github.com/vincent-petithory/dataurl"
)

type Config struct {
	DiscordToken        string            `required:"true"`
	DiscordWebhookId    string            `required:"true"`
	DiscordWebhookToken string            `required:"true"`
	ImgurClientId       string            `required:"true"`
	Servers             map[string]string `required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	imgur := &imgur.Client{
		ClientId: cfg.ImgurClientId,
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(listServers(cfg, imgur))
	dg.AddHandler(anyGamers(cfg))

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}

func anyGamers(cfg Config) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "any gamers") {
			return
		}

		embed, err := prepareEmbedWithServerStatus(cfg, nil, "pumpcraft", "mc.sep.gg")
		if err != nil {
			log.Printf("error preparing embed: %s", err.Error())
			return
		}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "gamers?",
			Embed:   embed,
		})
		if err != nil {
			log.Printf("error sending message: %s", err.Error())
			return
		}
	}
}

func prepareEmbedWithServerStatus(cfg Config, imgurClient *imgur.Client, serverName, host string) (*discordgo.MessageEmbed, error) {
	ctx := context.Background()
	hostports, err := resolveMinecraftHostPort(ctx, nil, host)
	if err != nil {
		return nil, fmt.Errorf("error resolving server host '%s': %s", host, err.Error())
	}
	if len(hostports) == 0 {
		return nil, nil
	}

	embed := &discordgo.MessageEmbed{
		Title: serverName,
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
		return embed, nil
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "host",
			Value: serverUrl,
		},
	}

	sample := make([]string, len(pong.Players.Sample))
	for i, player := range pong.Players.Sample {
		sample[i] = player.Name
	}
	playersEmbedField := &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("online (%d/%d)", pong.Players.Online, pong.Players.Max),
	}
	if len(sample) == 0 {
		playersEmbedField.Value = ":("
	} else {
		playersEmbedField.Value = strings.Join(sample, ", ")
	}
	embed.Fields = append(embed.Fields, playersEmbedField)
	embed.Color = 0x43b581

	embed.Description = pong.Description.Text

	if pong.Favicon != "" && imgurClient != nil {
		favIcon, err := dataurl.DecodeString(pong.Favicon)
		if err != nil {
			return nil, fmt.Errorf("error decoding favicon for server '%s': %s", serverUrl, err.Error())
		}

		uploadRequest := &imgur.ImageUploadRequest{
			Image: favIcon.Data,
			Name:  serverUrl,
		}
		img, err := imgurClient.Upload(ctx, uploadRequest)
		if err != nil {
			return nil, fmt.Errorf("error uploading favicon for server '%s' to imgur: %s", serverUrl, err.Error())
		}
		embed.Image = &discordgo.MessageEmbedImage{
			URL: img.Link,
		}
	}
	return embed, nil
}

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
