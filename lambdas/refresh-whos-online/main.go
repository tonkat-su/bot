package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers/connected"
	"github.com/tonkat-su/bot/handlers/refreshable"
)

type Config struct {
	MinecraftServerHost   string `split_words:"true" required:"true"`
	MinecraftServerName   string `split_words:"true" required:"true"`
	DiscordTokenSecretArn string `split_words:"true" required:"true"`
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

	secrets := secretsmanager.NewFromConfig(awsCfg)
	sv, err := secrets.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.DiscordTokenSecretArn),
	})
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

	lambda.Start(func() error {
		dg, err := discordgo.New("Bot " + *sv.SecretString)
		if err != nil {
			return err
		}
		dg.Identify.Compress = true

		closer := make(chan bool, 1)
		dg.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
			connectedHandler.OnConnect(s, event)
			log.Println("refreshed whos online")
			closer <- true
		})

		err = dg.Open()
		if err != nil {
			return err
		}

		<-closer
		err = dg.Close()
		if err != nil {
			return err
		}
		return nil
	})
}
