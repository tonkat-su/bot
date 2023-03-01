package interactions

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/state"
)

type Config struct {
	ImgurClientId        string `split_words:"true" required:"true"`
	DiscordToken         string `split_words:"true" required:"true"`
	DiscordWebhookPubkey string `split_words:"true" required:"true"`
	RconPassword         string `split_words:"true" required:"true"`
	RconHostport         string `split_words:"true" required:"true"`
}

func NewServer(cfg *Config) (*webhook.InteractionServer, error) {
	state := state.NewAPIOnlyState(cfg.DiscordToken, nil)
	r := &router{
		Router: cmdroute.NewRouter(),
		s:      state,
		cfg:    cfg,
	}

	/*imgurClient = &imgur.Client{
		ClientId: config.ImgurClientId,
	}
	*/

	r.Use(cmdroute.Deferrable(r.s, cmdroute.DeferOpts{}))

	r.AddFunc("whitelist", r.whitelist)
	r.AddFunc("online", r.online)

	return webhook.NewInteractionServer(cfg.DiscordWebhookPubkey, r)
}

type router struct {
	*cmdroute.Router
	s   *state.State
	cfg *Config
}

func (h *router) online(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return nil
}
