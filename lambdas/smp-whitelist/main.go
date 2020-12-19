package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/apex/gateway/v2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bsdlp/envconfig"
	"github.com/bsdlp/interactions"
)

type Config struct {
	RconPasswordSecretArn      string `required:"true" split_words:"true"`
	MinecraftServerRconAddress string `required:"true" split_words:"true"`
	DiscordApplicationPubkey   string `required:"true" split_words:"true"`
}

func decodePubkey(k string) (ed25519.PublicKey, error) {
	data, err := hex.DecodeString(k)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(data), nil
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	discordPubkey, err := decodePubkey(cfg.DiscordApplicationPubkey)
	if err != nil {
		log.Fatalf("error decoding discord pubkey: %s", err.Error())
	}

	lambda.Start(func(ctx context.Context, e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		req, err := gateway.NewRequest(ctx, e)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				Body:       err.Error(),
				StatusCode: http.StatusBadRequest,
			}, err
		}
		defer req.Body.Close()

		if req.Method != http.MethodPost {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusMethodNotAllowed,
			}, nil
		}

		verified := interactions.Verify(ctx, req, discordPubkey)
		if !verified {
			log.Println("invalid signature")
			return events.APIGatewayV2HTTPResponse{
				Body:       "invalid signature",
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		body, err := req.GetBody()
		if err != nil {
			log.Printf("error getting body from request: %s", err)
			return events.APIGatewayV2HTTPResponse{
				Body:       err.Error(),
				StatusCode: http.StatusInternalServerError,
			}, err
		}

		var data interactions.Data
		err = json.NewDecoder(body).Decode(&data)
		if err != nil {
			log.Println("invalid data")
			return events.APIGatewayV2HTTPResponse{
				Body:       "error decoding json request",
				StatusCode: http.StatusBadRequest,
			}, nil
		}

		if data.Type == interactions.Ping {
			log.Println("ping received")
			return events.APIGatewayV2HTTPResponse{
				Body:       `{"type":1}`,
				StatusCode: http.StatusOK,
			}, nil
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
		}, nil
	})
}
