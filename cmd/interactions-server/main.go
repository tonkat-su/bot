package main

import (
	"log"
	"net/http"

	"github.com/bsdlp/envconfig"
	"github.com/tonkat-su/bot/interactions"
)

var (
	config interactions.Config
	server *interactions.Server
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("error reading envconfig: %s", err.Error())
	}

	server, err = interactions.NewServer(&config)
	if err != nil {
		log.Fatalf("error initializing server: %s", err.Error())
	}
	defer func() {
		err := server.Close()
		if err != nil {
			log.Fatalf("error closing discord connection: %s", err.Error())
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/interactions", server)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Printf("error serving http server: %s", err.Error())
	}
}
