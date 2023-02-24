package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/bsdlp/envconfig"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/tonkat-su/bot/interactions"
)

type Config struct {
	ImgurClientId string `split_words:"true" required:"true"`

	// read from secrets manager using cdk https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_ecs.Secret.html
	DiscordToken         string `split_words:"true" required:"true"`
	DiscordWebhookUrl    string `split_words:"true" required:"true"`
	DiscordWebhookPubkey string `split_words:"true" required:"true"`
}

var (
	//imgurClient       *imgur.Client
	config            Config
	interactionServer *webhook.InteractionServer
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("error reading envconfig: %s", err.Error())
	}

	interactionServer, err = interactions.NewServer(&interactions.Config{
		ImgurClientId:        config.ImgurClientId,
		DiscordToken:         config.DiscordToken,
		DiscordWebhookPubkey: config.DiscordWebhookPubkey,
	})
	if err != nil {
		log.Fatalf("error initializing server: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/", interactionServer)

	lambda.Start(httpadapter.New(mux).ProxyWithContext)
}
