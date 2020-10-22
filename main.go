package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewtian/minepong"
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
	Servers             map[string]string `default:"hypixel;mc.hypixel.net,pumpcraft;mc.sep.gg" required:"true" kv_delimiter:";"`
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

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}

func listServers(cfg Config, imgurClient *imgur.Client) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content != "list servers" {
			return
		}

		ctx := context.Background()

		embeds := make([]*discordgo.MessageEmbed, 0, len(cfg.Servers))
		for serverName, host := range cfg.Servers {
			hostports, err := resolveMinecraftHostPort(ctx, nil, host)
			if err != nil {
				log.Printf("error resolving server host '%s': %s", host, err.Error())
				return
			}
			if len(hostports) == 0 {
				continue
			}
			serverUrl := hostports[0].String()

			conn, err := net.Dial("tcp", serverUrl)
			if err != nil {
				log.Printf("error connecting to server '%s': %s", serverUrl, err.Error())
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: serverName,
			}

			pong, err := minepong.Ping(conn, serverUrl)
			if err != nil {
				log.Printf("error pinging server '%s': %s", serverUrl, err.Error())
				embed.Fields = []*discordgo.MessageEmbedField{
					{
						Name:  "error",
						Value: err.Error(),
					},
				}
				embed.Color = 0xf04747
				embeds = append(embeds, embed)
				continue
			}

			embed.Fields = []*discordgo.MessageEmbedField{
				{
					Name:  "host",
					Value: host,
				},
				{
					Name:  "online",
					Value: fmt.Sprintf("%d/%d", pong.Players.Online, pong.Players.Max),
				},
			}
			embed.Color = 0x43b581

			if description, ok := pong.Description.(map[string]string); ok {
				embed.Description = description["text"]
			}

			if pong.FavIcon != "" {
				favIcon, err := dataurl.DecodeString(pong.FavIcon)
				if err != nil {
					log.Printf("error decoding favicon for server '%s': %s", serverUrl, err.Error())
					return
				}

				uploadRequest := &imgur.ImageUploadRequest{
					Image: favIcon.Data,
					Name:  serverUrl,
				}
				img, err := imgurClient.Upload(ctx, uploadRequest)
				if err != nil {
					log.Printf("error uploading favicon for server '%s' to imgur: %s", serverUrl, err.Error())
					return
				}
				embed.Image = &discordgo.MessageEmbedImage{
					URL: img.Link,
				}
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
