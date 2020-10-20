package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewtian/minepong"
	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DiscordToken string   `required:"true"`
	Servers      []string `default:"mc.sep.gg,mc.hypixel.net" required:"true"`
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

	dg.AddHandler(listServers)

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}

func listServers(cfg Config) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content != "list servers" {
			return
		}

		msg := &discordgo.MessageSend{}
		msg.Files = make([]*discordgo.File, 0, len(cfg.Servers))
		for _, host := range cfg.Servers {
			s, err := resolveMinecraftHostPort(context.Background(), nil, host)
			if err != nil {
				log.Printf("error resolving server host '%s': %s", host, err.Error())
			}
			if len(s) == 0 {
				continue
			}

			conn, err := net.Dial("tcp", s[0].String())
			if err != nil {
				log.Printf("error connecting to server '%s': %s", s[0].String(), err.Error())
			}
			pong, err := minepong.Ping(conn, s[0].String())
			if err != nil {
				log.Printf("error pinging server '%s': %s", s[0].String(), err.Error())
			}

			file, err := parseFavIcon(s[0].String(), pong.FavIcon)
			if err != nil {
				log.Printf("error parsing favicon: %s", err.Error())
			}
			msg.Files = append(msg.Files, file)
		}
		_, err := s.ChannelMessageSendComplex(m.ChannelID, msg)
		if err != nil {
			log.Printf("error responding to list servers: %s", err.Error())
		}
	}
}
