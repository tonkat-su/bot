package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	mcrcon "github.com/Kelwing/mc-rcon"
	"github.com/apex/gateway/v2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bsdlp/discord-interactions-go/interactions"
	"github.com/bsdlp/envconfig"
)

type Config struct {
	RconPasswordSecretArn      string `required:"true" split_words:"true"`
	MinecraftServerRconAddress string `required:"true" split_words:"true"`
	DiscordApplicationPubkey   string `required:"true" split_words:"true"`
	rconPassword               string
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

	// retrieve rcon password from aws secrets manager
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("error loading aws config: %s", err)
	}
	sv, err := secretsmanager.NewFromConfig(awsCfg).GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.RconPasswordSecretArn),
	})
	if err != nil {
		log.Fatal(err)
	}
	cfg.rconPassword = *sv.SecretString

	// hex decode discord pubkey
	discordPubkey, err := decodePubkey(cfg.DiscordApplicationPubkey)
	if err != nil {
		log.Fatalf("error decoding discord pubkey: %s", err.Error())
	}

	// register lambda function
	lambda.Start(func(ctx context.Context, e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		// convert a proxied api gateway v2 http request into an *http.Request
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

		// validate signature from discord and return 401 if invalid
		verified := interactions.Verify(req, discordPubkey)
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

		// marshal interaction webhook data
		var data interactions.Data
		err = json.NewDecoder(body).Decode(&data)
		if err != nil {
			log.Printf("invalid data: %s", err.Error())
			return events.APIGatewayV2HTTPResponse{
				Body:       "error decoding json request",
				StatusCode: http.StatusBadRequest,
			}, err
		}

		// reply with a pong when discord pings us
		if data.Type == interactions.Ping {
			log.Println("ping received")
			return events.APIGatewayV2HTTPResponse{
				Body:       `{"type":1}`,
				StatusCode: http.StatusOK,
			}, nil
		}

		// actual handler
		return handle(&cfg, data)
	})
}

func handle(cfg *Config, data interactions.Data) (events.APIGatewayV2HTTPResponse, error) {
	conn := &mcrcon.MCConn{}
	err := conn.Open(cfg.MinecraftServerRconAddress, cfg.rconPassword)
	if err != nil {
		log.Printf("unable to connect: %s", err.Error())
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusFailedDependency,
			Body:       err.Error(),
		}, err
	}
	defer conn.Close()

	err = conn.Authenticate()
	if err != nil {
		log.Printf("unable to authenticate: %s", err.Error())
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusFailedDependency,
			Body:       "unable to authenticate by rcon",
		}, err
	}

	var rconCommand string
	subcommand := data.Data.Options[0].Options[0]
	switch subcommand.Name {
	case "add":
		for _, v := range subcommand.Options {
			if v.Name == "mc_user" {
				if username, ok := v.Value.(string); ok {
					rconCommand = fmt.Sprintf("whitelist add %s", username)
				}
			}
		}
	case "list":
		rconCommand = "whitelist list"
	case "remove":
		for _, v := range subcommand.Options {
			if v.Name == "mc_user" {
				if username, ok := v.Value.(string); ok {
					rconCommand = fmt.Sprintf("whitelist remove %s", username)
				}
			}
		}
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       fmt.Sprintf("unrecognized whitelist subcommand: %s", subcommand.Name),
		}, nil
	}

	if rconCommand == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       errMinecraftUsernameRequired.Error(),
		}, errMinecraftUsernameRequired
	}

	log.Printf("sending rcon command: %s", rconCommand)

	resp, err := conn.SendCommand(rconCommand)
	if err != nil {
		log.Printf("error sending command: %s", err.Error())
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusFailedDependency,
			Body:       err.Error(),
		}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       resp,
	}, nil
}

var errMinecraftUsernameRequired = errors.New("missing username parameter")
