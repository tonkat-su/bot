package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mcrcon "github.com/Kelwing/mc-rcon"
	"github.com/apex/gateway/v2"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// validate signature from discord and return 401 if invalid
		verified := interactions.Verify(r, discordPubkey)
		if !verified {
			log.Println("invalid signature")
			writeResponse(w, http.StatusUnauthorized, "invalid signature")
			return
		}

		// marshal interaction webhook data
		var data interactions.Data
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Printf("invalid data: %s", err.Error())
			writeResponse(w, http.StatusBadRequest, "error decoding json request")
			return
		}

		// reply with a pong when discord pings us
		if data.Type == interactions.Ping {
			log.Println("ping received")
			writeResponse(w, http.StatusOK, `{"type":1}`)
			return
		}

		conn := &mcrcon.MCConn{}
		err := conn.Open(cfg.MinecraftServerRconAddress, cfg.rconPassword)
		if err != nil {
			log.Printf("unable to connect: %s", err.Error())
			writeResponse(w, http.StatusFailedDependency, err.Error())
			return
		}
		defer conn.Close()

		err = conn.Authenticate()
		if err != nil {
			log.Printf("unable to authenticate: %s", err.Error())
			writeResponse(w, http.StatusFailedDependency, "unable to authenticate by rcon")
			return
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
			writeResponse(w, http.StatusUnprocessableEntity, fmt.Sprintf("unrecognized whitelist subcommand: %s", subcommand.Name))
			return
		}

		if rconCommand == "" {
			writeResponse(w, http.StatusUnprocessableEntity, "missing minecraft username parameter")
			return
		}

		log.Printf("sending rcon command: %s", rconCommand)

		resp, err := conn.SendCommand(rconCommand)
		if err != nil {
			log.Printf("error sending command: %s", err.Error())
			writeResponse(w, http.StatusFailedDependency, err.Error())
			return
		}

		log.Printf("response: %s", resp)
		writeResponse(w, http.StatusOK, resp)
	})

	log.Fatal(gateway.ListenAndServe(":3000", nil))
}

func writeResponse(w http.ResponseWriter, statusCode int, body string) {
	err := json.NewEncoder(w).Encode(interactions.InteractionResponse{
		Type: 4,
		Data: &interactions.InteractionApplicationCommandCallbackData{
			Content: body,
		},
	})
	if err != nil {
		log.Printf("failed to encode body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
}
