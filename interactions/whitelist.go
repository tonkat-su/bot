package interactions

import (
	"context"
	"fmt"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/jltobler/go-rcon"
)

func (h *router) whitelist(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	rconClient := rcon.NewClient("rcon://"+h.cfg.RconHostport, h.cfg.RconPassword)

	var rconCommand string
	switch cmd.Name {
	case "list":
		rconCommand = "whitelist list"
	case "add":
		option := cmd.Options.Find("username")
		username := option.String()
		rconCommand = fmt.Sprintf("whitelist add %s", username)
	case "remove":
		option := cmd.Options.Find("username")
		username := option.String()
		rconCommand = fmt.Sprintf("whitelist remove %s", username)
	default:
		return &api.InteractionResponseData{
			Content: option.NewNullableString("invalid command"),
		}
	}

	output, err := rconClient.Send(rconCommand)
	if err != nil {
		log.Printf("error sending rcon command: %s", err.Error())
		return &api.InteractionResponseData{
			Content:         option.NewNullableString("error sending rcon command"),
			Flags:           discord.EphemeralMessage,
			AllowedMentions: &api.AllowedMentions{},
		}
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString(output),
	}
}
