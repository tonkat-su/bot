package connected

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/v2/emoji"
	"github.com/tonkat-su/bot/v2/handlers"
	"github.com/tonkat-su/bot/v2/imgur"
	"github.com/tonkat-su/bot/v2/mclookup"
	"github.com/tonkat-su/bot/v2/mcuser"
	"github.com/vincent-petithory/dataurl"
)

type RefreshableBackend struct {
	MinecraftServerHost string
	MinecraftServerName string
	Imgur               *imgur.Client
}

func ReplyWithServerStatus(host, name string, imgurClient *imgur.Client) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !(strings.HasPrefix(m.Content, "any gamers") || strings.HasPrefix(m.Content, "Any gamers")) {
			return
		}

		embed, err := prepareStatusEmbed(s, m.GuildID, host, name, imgurClient)
		if err != nil {
			log.Printf("error preparing status embed: %s", err)
			return
		}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "gamers?",
			Embed:   embed,
		})
		if err != nil {
			log.Printf("error replying with server status: %s", err)
			return
		}
	}
}

func prepareStatusEmbed(s *discordgo.Session, guildID, host, name string, imgurClient *imgur.Client) (*discordgo.MessageEmbed, error) {
	ctx := context.Background()
	hostports, err := mclookup.ResolveMinecraftHostPort(ctx, nil, host)
	if err != nil {
		return nil, fmt.Errorf("error resolving server host '%s': %s", host, err.Error())
	}
	if len(hostports) == 0 {
		return nil, nil
	}

	embed := &discordgo.MessageEmbed{
		Title: name,
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
		err = emoji.SyncMinecraftAvatarsToEmoji(s, guildID, players)
		if err != nil {
			log.Printf("error syncing avatars to emoji: %s", err)
		}

		playersEmbedField = &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("online (%d/%d)", pong.Players.Online, pong.Players.Max),
			Value: emoji.PlayerListEmojis(players),
		}
	}
	embed.Fields = append(embed.Fields, playersEmbedField)

	updatedFields, err := handlers.AppendLastUpdatedEmbedField(embed.Fields, lastUpdated)
	if err != nil {
		log.Printf("error appending last updated embed field: %s", err)
	} else {
		embed.Fields = updatedFields
	}

	embed.Color = 0x43b581

	embed.Description = pong.Description.Text

	if pong.Favicon != "" && imgurClient != nil {
		favIcon, err := dataurl.DecodeString(pong.Favicon)
		if err != nil {
			return nil, fmt.Errorf("error decoding favicon for server '%s': %s", serverUrl, err.Error())
		}

		uploadRequest := &imgur.ImageUploadRequest{
			Image: favIcon.Data,
			Name:  serverUrl,
		}
		img, err := imgurClient.Upload(ctx, uploadRequest)
		if err != nil {
			return nil, fmt.Errorf("error uploading favicon for server '%s' to imgur: %s", serverUrl, err.Error())
		}
		embed.Image = &discordgo.MessageEmbedImage{
			URL: img.Link,
		}
	}
	return embed, nil
}

func (h *RefreshableBackend) CreateRefreshableMessage(s *discordgo.Session, guildID string, channelID string) (*discordgo.Message, error) {
	embed, err := prepareStatusEmbed(s, guildID, h.MinecraftServerHost, h.MinecraftServerName, h.Imgur)
	if err != nil {
		return nil, err
	}

	return s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "gamers currently online",
		Embed:   embed,
	})
}

func (h *RefreshableBackend) RefreshMessage(s *discordgo.Session, event *discordgo.MessageReaction) error {
	embed, err := prepareStatusEmbed(s, event.GuildID, h.MinecraftServerHost, h.MinecraftServerName, h.Imgur)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content: aws.String("gamers currently online"),
		Embed:   embed,
		ID:      event.MessageID,
		Channel: event.ChannelID,
	})
	return err
}
