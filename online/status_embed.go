package online

import (
	"context"
	"fmt"
	"log"
	"strings"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/emoji"
	"github.com/tonkat-su/bot/mclookup"
)

type PrepareStatusRequest struct {
	Session *discordgo.Session

	GuildId        string
	ServerHostname string
	ServerName     string
}

type PrepareStatusResponse struct {
	MessageEmbeds []*discordgo.MessageEmbed
	Files         []*discordgo.File
}

func PrepareStatus(params *PrepareStatusRequest) (*PrepareStatusResponse, error) {
	ctx := context.Background()
	hostports, err := mclookup.ResolveMinecraftHostPort(ctx, nil, params.ServerHostname)
	if err != nil {
		return nil, fmt.Errorf("error resolving server host '%s': %s", params.ServerHostname, err.Error())
	}
	if len(hostports) == 0 {
		return nil, nil
	}

	embed := &discordgo.MessageEmbed{
		Title: params.ServerName,
	}

	serverUrl := hostports[0].String()
	pong, err := mcpinger.New(hostports[0].Host, hostports[0].Port).Ping()
	if err != nil {
		log.Printf("error pinging server '%s': %s", serverUrl, err.Error())
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "error",
				Value: err.Error(),
			},
		}
		embed.Color = 0xf04747
		return &PrepareStatusResponse{
			MessageEmbeds: []*discordgo.MessageEmbed{embed},
		}, nil
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "host",
			Value: serverUrl,
		},
	}

	players := make([]*emoji.Player, len(pong.Players.Sample))
	for i, p := range pong.Players.Sample {
		players[i] = &emoji.Player{
			Name: p.Name,
		}
	}

	var playersEmbedField *discordgo.MessageEmbedField
	if len(players) == 0 {
		playersEmbedField = &discordgo.MessageEmbedField{
			Name:  "nobody's online :(",
			Value: "https://www.youtube.com/watch?v=ypVpv-fEevk",
		}
	} else {
		// fill emoji ids for players
		err = emoji.HydrateEmojiIds(params.Session, params.GuildId, players)
		if err != nil {
			log.Printf("error syncing avatars to emoji: %s", err)
		}

		// format into list of face emojis of online players
		emojis := make([]string, len(players))
		for i, p := range players {
			emojis[i] = p.EmojiTextCode()
		}
		emojiString := strings.Join(emojis, " ")

		playersEmbedField = &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("online (%d/%d)", pong.Players.Online, pong.Players.Max),
			Value: emojiString,
		}
	}
	embed.Fields = append(embed.Fields, playersEmbedField)

	embed.Color = 0x43b581

	embed.Description = pong.Description.Text

	return &PrepareStatusResponse{
		MessageEmbeds: []*discordgo.MessageEmbed{embed},
	}, nil
}
