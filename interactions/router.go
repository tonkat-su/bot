package interactions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type Config struct {
	ImgurClientId        string
	DiscordToken         string
	DiscordWebhookPubkey string
}

func NewServer(cfg *Config) (*webhook.InteractionServer, error) {
	state := state.NewAPIOnlyState(cfg.DiscordToken, nil)
	r := &router{
		Router: cmdroute.NewRouter(),
		s:      state,
	}

	/*imgurClient = &imgur.Client{
		ClientId: config.ImgurClientId,
	}
	*/

	r.Use(cmdroute.Deferrable(r.s, cmdroute.DeferOpts{}))

	r.AddFunc("initialize", r.initializeMessage)
	r.AddFunc("refresh", r.refreshMessage)

	return webhook.NewInteractionServer(cfg.DiscordWebhookPubkey, r)
}

type router struct {
	*cmdroute.Router
	s *state.State
}

type msgMeta struct {
	ServerName string
	HostPort   string
}

func (h *router) initializeMessage(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
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

func (h *router) refreshMessage(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return nil
}
