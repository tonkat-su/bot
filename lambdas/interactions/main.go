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

var (
	//imgurClient       *imgur.Client
	config            interactions.Config
	interactionServer *webhook.InteractionServer
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("error reading envconfig: %s", err.Error())
	}

	interactionServer, err = interactions.NewServer(&config)
	if err != nil {
		log.Fatalf("error initializing server: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/interactions", interactionServer)

	lambda.Start(httpadapter.New(mux).ProxyWithContext)
}
