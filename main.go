package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	DiscordToken        string `required:"true" split_words:"true"`
	MinecraftServerName string `required:"true" split_words:"true"`
	MinecraftServerHost string `required:"true" split_words:"true"`
	GuildId             string `required:"true" split_words:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

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

	presenceTicker := time.NewTicker(5 * time.Minute)
	go func(presenceTicker *time.Ticker) {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := updatePresence(ctx, dg, cfg)
			if err != nil {
				log.Printf("failed to update presence: %s", err.Error())
			}
			cancel()
			<-presenceTicker.C
		}
	}(presenceTicker)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}
