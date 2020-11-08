package main

import (
	"context"
	"fmt"
	"log"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/leaderboard"
)

func updatePresence(ctx context.Context, cfg Config, s *discordgo.Session, lboard *leaderboard.Service) error {
	hostports, err := resolveMinecraftHostPort(ctx, nil, cfg.MinecraftServerHost)
	if err != nil {
		return fmt.Errorf("error resolving server host '%s': %s", cfg.MinecraftServerHost, err.Error())
	}
	if len(hostports) == 0 {
		log.Printf("no records for %s", cfg.MinecraftServerHost)
		return nil
	}

	pong, err := mcpinger.New(hostports[0].Host, hostports[0].Port).Ping()
	if err != nil {
		return err
	}

	if pong.Players.Online > 0 {
		go func(pong *mcpinger.ServerInfo) {
			input := &leaderboard.RecordScoresInput{
				Scores: make([]*leaderboard.PlayerScore, len(pong.Players.Sample)),
			}
			for i, v := range pong.Players.Sample {
				input.Scores[i] = &leaderboard.PlayerScore{
					PlayerId: v.ID,
					Score:    cfg.PresenceInterval,
				}
			}
			err = lboard.RecordScores(context.TODO(), input)
			if err != nil {
				log.Printf("error recording score for leaderboard: %s", err)
			}
		}(pong)

		return s.UpdateStatus(0, fmt.Sprintf("currently online: (%d/%d)", pong.Players.Online, pong.Players.Max))
	}
	return s.UpdateStatus(0, "")
}
