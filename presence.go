package main

import (
	"context"
	"fmt"
	"log"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
)

func updatePresence(ctx context.Context, s *discordgo.Session, cfg Config) error {
	hostports, err := resolveMinecraftHostPort(ctx, nil, cfg.ServerHost)
	if err != nil {
		return fmt.Errorf("error resolving server host '%s': %s", cfg.ServerHost, err.Error())
	}
	if len(hostports) == 0 {
		log.Printf("no records for %s", cfg.ServerHost)
		return nil
	}

	pong, err := mcpinger.New(hostports[0].Host, hostports[0].Port).Ping()
	if err != nil {
		return err
	}

	if pong.Players.Online > 0 {
		return s.UpdateStatus(0, fmt.Sprintf("currently online: (%d/%d)", pong.Players.Online, pong.Players.Max))
	}
	return s.UpdateStatus(0, "")
}
