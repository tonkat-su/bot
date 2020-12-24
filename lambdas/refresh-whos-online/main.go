package main

import (
	"context"
	"log"

	"github.com/andersfylling/disgord"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers/connected"
	"github.com/tonkat-su/bot/handlers/refreshable"
)

type config struct {
	MinecraftServerHost   string `split_words:"true" required:"true"`
	MinecraftServerName   string `split_words:"true" required:"true"`
	DiscordTokenSecretArn string `split_words:"true" required:"true"`
}

func main() {
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	secrets := secretsmanager.New(sess)
	sv, err := secrets.GetSecretValueWithContext(context.TODO(), &secretsmanager.GetSecretValueInput{
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
		dg := disgord.New(disgord.Config{
			BotToken: aws.StringValue(sv.SecretString),
		})
		defer dg.Gateway().StayConnectedUntilInterrupted()

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
