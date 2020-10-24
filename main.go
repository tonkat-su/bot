package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	DiscordToken        string `required:"true"`
	DiscordWebhookId    string `required:"true"`
	DiscordWebhookToken string `required:"true"`
	ImgurClientId       string `required:"true"`
	ServerName          string `required:"true"`
	ServerHost          string `required:"true"`
	GuildId             string `required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	/*
		imgur := &imgur.Client{
			ClientId: cfg.ImgurClientId,
		}
	*/

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}
	dg.ShouldReconnectOnError = true
	dg.StateEnabled = true
	dg.Identify.Compress = true

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
