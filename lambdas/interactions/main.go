package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/bsdlp/envconfig"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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

type interactionHandler struct {
	*cmdroute.Router
	s *state.State
}

type msgMeta struct {
	ServerName string
	HostPort   string
}

func (h *interactionHandler) initializeMessage(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	var meta msgMeta
	err := json.NewDecoder(strings.NewReader(cmd.Event.Message.Content)).Decode(&meta)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("error decoding meta: " + err.Error()),
		}
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString(fmt.Sprintf("servername: %s, hostport: %s", meta.ServerName, meta.HostPort)),
	}
}

func (h *interactionHandler) refreshMessage(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return nil
}

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("error reading envconfig: %s", err.Error())
	}

	/*imgurClient = &imgur.Client{
		ClientId: config.ImgurClientId,
	}
	*/
	state := state.NewAPIOnlyState(config.DiscordToken, nil)

	h := &interactionHandler{
		s:      state,
		Router: cmdroute.NewRouter(),
	}
	h.Use(cmdroute.Deferrable(h.s, cmdroute.DeferOpts{}))
	h.AddFunc("initialize", h.initializeMessage)
	h.AddFunc("refresh", h.refreshMessage)

	interactionServer, err = webhook.NewInteractionServer(config.DiscordWebhookPubkey, h)
	if err != nil {
		log.Fatalf("error creating interaction server: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/", interactionServer)

	lambda.Start(httpadapter.New(mux).ProxyWithContext)
}