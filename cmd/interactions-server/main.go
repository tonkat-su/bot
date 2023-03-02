package main

import (
	"log"
	"net/http"

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

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Printf("error serving http server: %s", err.Error())
	}
}
