package main

import (
	"context"
	"fmt"
	"log"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bsdlp/envconfig"
	"github.com/tonkat-su/bot/leaderboard"
	"github.com/tonkat-su/bot/mclookup"
)

// triggered by cloudwatch event to query the minecraft server and give cat treats to players
func Handler(cfg Config, leaderboardService *leaderboard.Service) func(context.Context, *events.CloudWatchEvent) error {
	return func(ctx context.Context, event *events.CloudWatchEvent) error {
		hostports, err := mclookup.ResolveMinecraftHostPort(ctx, nil, cfg.MinecraftServerHost)
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

		if pong.Players.Online == 0 {
			return nil
		}

		input := &leaderboard.RecordScoresInput{
			Scores: make([]*leaderboard.PlayerScore, len(pong.Players.Sample)),
		}
		for i, v := range pong.Players.Sample {
			input.Scores[i] = &leaderboard.PlayerScore{
				PlayerId: v.ID,
				Score:    1,
			}
		}
		return leaderboardService.RecordScores(ctx, input)
	}
}

type Config struct {
	MinecraftServerName string `required:"true" split_words:"true"`
	MinecraftServerHost string `required:"true" split_words:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	leaderboardService, err := leaderboard.New(session, &leaderboard.Config{
		NamespacePrefix: cfg.MinecraftServerName,
	})
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(Handler(cfg, leaderboardService))
}
