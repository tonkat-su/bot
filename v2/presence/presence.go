package presence

import (
	"context"
	"fmt"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/v2/mclookup"
)

func Update(ctx context.Context, host string, s *discordgo.Session) error {
	hostports, err := mclookup.ResolveMinecraftHostPort(ctx, nil, host)
	if err != nil {
		return fmt.Errorf("error resolving server host '%s': %s", host, err.Error())
	}
	if len(hostports) == 0 {
		return s.UpdateStatus(0, "")
	}

	pong, err := mcpinger.New(hostports[0].Host, hostports[0].Port).Ping()
	if err != nil {
		return err
	}

	if pong.Players.Online > 0 {
		return s.UpdateStatus(0, fmt.Sprintf("currently online: (%d/%d)", pong.Players.Online, pong.Players.Max))
	}
	return s.UpdateStatus(0, "")
}
