package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers/connected"
	"github.com/tonkat-su/bot/handlers/echo"
	"github.com/tonkat-su/bot/handlers/pinnedleaderboard"
	"github.com/tonkat-su/bot/handlers/refreshable"
	"github.com/tonkat-su/bot/handlers/register"
	"github.com/tonkat-su/bot/leaderboard"
	"github.com/tonkat-su/bot/presence"
	"github.com/tonkat-su/bot/users"
)

type Config struct {
	DiscordToken       string `required:"true" split_words:"true"`
	AWSRegion          string `required:"true" envconfig:"AWS_REGION"`
	AWSAccessKeyId     string `required:"true" envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `required:"true" envconfig:"AWS_SECRET_ACCESS_KEY"`

	MinecraftServerName string `required:"true" split_words:"true"`
	MinecraftServerHost string `required:"true" split_words:"true"`
	GuildId             string `required:"true" split_words:"true"`
	PresenceInterval    string `default:"5m" split_words:"true"`

	UsersServiceTableName string `default:"TonkatsuStack-users9E3E6EF7-19OQ46A0WAOHQ" split_words:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("error loading aws config: %s", err)
	}

	usersService, err := users.New(awsCfg, cfg.UsersServiceTableName)
	if err != nil {
		log.Fatalf("error setting up users service: %s", err)
	}

	leaderboardService, err := leaderboard.New(awsCfg, &leaderboard.Config{NamespacePrefix: cfg.MinecraftServerName})
	if err != nil {
		log.Fatalf("error setting up leaderboard service: %s", err)
	}

	refreshableLeaderboard := &refreshable.Handler{
		Backend: &pinnedleaderboard.RefreshableBackend{
			Leaderboard: leaderboardService,
		},
		PinnedChannelName: "leaderboard",
	}

	whosConnected := &refreshable.Handler{
		Backend: &connected.RefreshableBackend{
			MinecraftServerName: cfg.MinecraftServerName,
			MinecraftServerHost: cfg.MinecraftServerHost,
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

	dg.AddHandler(echo.Echo)
	dg.AddHandler(register.RegisterMinecraftGamer(usersService))
	dg.AddHandler(register.LookupUser(usersService))
	refreshableLeaderboard.AddHandlers(dg)
	whosConnected.AddHandlers(dg)

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		interval, err := time.ParseDuration(cfg.PresenceInterval)
		if err != nil {
			log.Printf("invalid presence interval: %s", err)
			interval = 5 * time.Minute
		}

		presenceTicker := time.NewTicker(interval)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := presence.Update(ctx, cfg.MinecraftServerHost, dg)
			if err != nil {
				log.Printf("failed to update presence: %s", err.Error())
			}
			cancel()
			<-presenceTicker.C
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-sc
	dg.Close()
}
