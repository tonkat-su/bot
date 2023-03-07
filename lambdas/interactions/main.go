package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
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
	mux.HandleFunc("/keepwarm", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	lambda.Start(httpadapter.New(mux).ProxyWithContext)
}
