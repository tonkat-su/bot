package interactions

import (
	"context"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/jltobler/go-rcon"
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

	r.AddFunc("ping", r.ping)
	r.AddFunc("whitelist", r.whitelist)
	r.AddFunc("online", r.online)
	r.AddFunc("echo", r.echo)

	return webhook.NewInteractionServer(cfg.DiscordWebhookPubkey, r)
}

type router struct {
	*cmdroute.Router
	s   *state.State
	cfg *Config
}

func (h *router) online(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	rconClient := rcon.NewClient("rcon://"+h.cfg.RconHostport, h.cfg.RconPassword)
	output, err := rconClient.Send("list")
	if err != nil {
		log.Printf("error sending list command: %s", err.Error())
	}
	return &api.InteractionResponseData{
		Content: option.NewNullableString(output),
	}
}

func (h *router) ping(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
	}

}

func (h *router) echo(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	var options struct {
		Arg string `discord:"message"`
	}

	if err := cmd.Options.Unmarshal(&options); err != nil {
		log.Printf("error unmarshaling echo: %s", err.Error())
	}

	return &api.InteractionResponseData{
		Content:         option.NewNullableString(options.Arg),
		AllowedMentions: &api.AllowedMentions{},
	}
}
