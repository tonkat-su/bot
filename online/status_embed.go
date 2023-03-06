package online

import (
	"context"
	"fmt"
	"log"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/emoji"
	"github.com/tonkat-su/bot/imgur"
	"github.com/tonkat-su/bot/mclookup"
	"github.com/vincent-petithory/dataurl"
)

type PrepareStatusEmbedRequest struct {
	Session *discordgo.Session
	Imgur   *imgur.Client

	GuildId        string
	ServerHostname string
	ServerName     string
}

func PrepareStatusEmbed(params *PrepareStatusEmbedRequest) (*discordgo.MessageEmbed, error) {
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
		return embed, nil
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
		err = emoji.SyncMinecraftAvatarsToEmoji(params.Session, params.GuildId, players)
		if err != nil {
			log.Printf("error syncing avatars to emoji: %s", err)
		}

		playersEmbedField = &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("online (%d/%d)", pong.Players.Online, pong.Players.Max),
			Value: emoji.PlayerListEmojis(players),
		}
	}
	embed.Fields = append(embed.Fields, playersEmbedField)

	embed.Color = 0x43b581

	embed.Description = pong.Description.Text

	if pong.Favicon != "" && params.Imgur != nil {
		favIcon, err := dataurl.DecodeString(pong.Favicon)
		if err != nil {
			return nil, fmt.Errorf("error decoding favicon for server '%s': %s", serverUrl, err.Error())
		}

		uploadRequest := &imgur.ImageUploadRequest{
			Image: favIcon.Data,
			Name:  serverUrl,
		}
		img, err := params.Imgur.Upload(ctx, uploadRequest)
		if err != nil {
			return nil, fmt.Errorf("error uploading favicon for server '%s' to imgur: %s", serverUrl, err.Error())
		}
		embed.Image = &discordgo.MessageEmbedImage{
			URL: img.Link,
		}
	}
	return embed, nil
}
