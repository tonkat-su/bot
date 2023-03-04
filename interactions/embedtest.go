package interactions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/emoji"
	"github.com/tonkat-su/bot/handlers"
	"github.com/tonkat-su/bot/imgur"
	"github.com/tonkat-su/bot/mclookup"
	"github.com/tonkat-su/bot/mcuser"
	"github.com/vincent-petithory/dataurl"
)

func (srv *Server) test(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	messageEmbed, err := prepareStatusEmbed(&prepareStatusEmbedRequest{
		Session:                     s,
		Imgur:                       srv.imgur,
		GuildId:                     srv.cfg.DiscordGuildId,
		ServerHostname:              srv.cfg.MinecraftServerHost,
		ServerName:                  srv.cfg.MinecraftServerName,
		AppendLastUpdatedEmbedField: false,
	})
	if err != nil {
		log.Printf("error rendering online message embed: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = s.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
		Content: "gamers currently online",
		Embeds:  []*discordgo.MessageEmbed{messageEmbed},
	})
	if err != nil {
		log.Printf("error sending message: %s", err)
	}

	response := discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "<:tigglywuffFace:1081372739891372093>",
			Embeds:  []*discordgo.MessageEmbed{messageEmbed},
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}

	err = replyToInteraction(event, response)
	if err != nil {
		log.Printf("failed to encode body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type prepareStatusEmbedRequest struct {
	Session *discordgo.Session
	Imgur   *imgur.Client

	GuildId                     string
	ServerHostname              string
	ServerName                  string
	AppendLastUpdatedEmbedField bool
}

func prepareStatusEmbed(params *prepareStatusEmbedRequest) (*discordgo.MessageEmbed, error) {
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

	lastUpdated := time.Now()

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "host",
			Value: serverUrl,
		},
	}

	players := make([]*emoji.Player, len(pong.Players.Sample))
	for i, p := range pong.Players.Sample {
		uuid, err := mcuser.GetUuid(p.Name)
		if err != nil {
			uuid = p.ID
			log.Printf("error getting uuid for user %s: %s", p.Name, err.Error())
		}
		players[i] = &emoji.Player{
			Name: p.Name,
			Uuid: uuid,
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
			Value: emoji.PlayerListEmojis([]*emoji.Player{players[0]}),
		}
	}
	embed.Fields = append(embed.Fields, playersEmbedField)

	if params.AppendLastUpdatedEmbedField {
		updatedFields, err := handlers.AppendLastUpdatedEmbedField(embed.Fields, lastUpdated)
		if err != nil {
			log.Printf("error appending last updated embed field: %s", err)
		} else {
			embed.Fields = updatedFields
		}
	}

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
