package main

import (
	"log"

	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers/connected"
	"github.com/tonkat-su/bot/handlers/refreshable"
)

type config struct {
	MinecraftServerHost string `split_words:"true" required:"true"`
	MinecraftServerName string `split_words:"true" required:"true"`
	GuildId             string `split_words:"true" required:"true"`
	DiscordToken        string `split_words:"true" required:"true"`
}

func main() {
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	connectedHandler := &refreshable.Handler{
		Backend: &connected.RefreshableBackend{
			MinecraftServerHost: cfg.MinecraftServerHost,
			MinecraftServerName: cfg.MinecraftServerName,
		},
		PinnedChannelName: "whos-online",
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}

	dg.ShouldReconnectOnError = true
	dg.StateEnabled = true
	dg.Identify.Compress = true

	closer := make(chan bool, 1)
	dg.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		connectedHandler.OnConnect(s, event)
		closer <- true
	})

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	<-closer
	dg.Close()
}
