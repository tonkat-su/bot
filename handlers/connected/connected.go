package connected

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/imgur"
	"github.com/tonkat-su/bot/online"
)

type RefreshableBackend struct {
	MinecraftServerHost string
	MinecraftServerName string
	Imgur               *imgur.Client
}

func (h *RefreshableBackend) CreateRefreshableMessage(s *discordgo.Session, guildID string, channelID string) (*discordgo.Message, error) {
	embed, err := online.PrepareStatusEmbed(&online.PrepareStatusEmbedRequest{
		Session:                     s,
		Imgur:                       h.Imgur,
		GuildId:                     guildID,
		ServerHostname:              h.MinecraftServerHost,
		ServerName:                  h.MinecraftServerName,
		AppendLastUpdatedEmbedField: true,
	})
	if err != nil {
		return nil, err
	}

	return s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "gamers currently online",
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
}

func (h *RefreshableBackend) RefreshMessage(s *discordgo.Session, event *discordgo.MessageReaction) error {
	embed, err := online.PrepareStatusEmbed(&online.PrepareStatusEmbedRequest{
		Session:                     s,
		Imgur:                       h.Imgur,
		GuildId:                     event.GuildID,
		ServerHostname:              h.MinecraftServerHost,
		ServerName:                  h.MinecraftServerName,
		AppendLastUpdatedEmbedField: true,
	})
	if err != nil {
		return err
	}

	embeds := make([]*discordgo.MessageEmbed, 0, 1)
	if embed != nil {
		embeds = append(embeds, embed)
	}
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content: aws.String("gamers currently online"),
		Embeds:  embeds,
		ID:      event.MessageID,
		Channel: event.ChannelID,
	})
	return err
}
