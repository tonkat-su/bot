package pinnedleaderboard

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers"
	"github.com/tonkat-su/bot/leaderboard"
	"github.com/tonkat-su/bot/mcuser"
)

type PrepareStandingsEmbedRequest struct {
	Standings                   *leaderboard.Standings
	AppendLastUpdatedEmbedField bool
}

func PrepareStandingsEmbed(params *PrepareStandingsEmbedRequest) (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: `biggest nerds on the server
(in the last 7 days)`,
		Fields: make([]*discordgo.MessageEmbedField, len(params.Standings.SortedStandings), len(params.Standings.SortedStandings)+1),
	}
	for i, v := range params.Standings.SortedStandings {
		username, err := mcuser.GetUsername(v.PlayerId)
		if err != nil {
			return nil, fmt.Errorf("error getting username: %s", err)
		}

		line := "%d cat treats"
		if v.Score == 1 {
			line = "%d cat treat"
		}
		embed.Fields[i] = &discordgo.MessageEmbedField{
			Name:  username,
			Value: fmt.Sprintf(line, v.Score),
		}
	}

	if params.AppendLastUpdatedEmbedField {
		var err error
		embed.Fields, err = handlers.AppendLastUpdatedEmbedField(embed.Fields, params.Standings.LastUpdated)
		if err != nil {
			return nil, err
		}
	}

	return embed, nil
}

func sendStandingsMessage(s *discordgo.Session, channelID string, standings *leaderboard.Standings) (*discordgo.Message, error) {
	embed, err := PrepareStandingsEmbed(&PrepareStandingsEmbedRequest{
		Standings:                   standings,
		AppendLastUpdatedEmbedField: true,
	})
	if err != nil {
		return nil, err
	}
	return s.ChannelMessageSendEmbed(channelID, embed)
}

type RefreshableBackend struct {
	Leaderboard *leaderboard.Service
}

func (h *RefreshableBackend) CreateRefreshableMessage(s *discordgo.Session, guildID string, channelID string) (*discordgo.Message, error) {
	standings, err := h.Leaderboard.GetStandings(context.TODO())
	if err != nil {
		return nil, err
	}
	return sendStandingsMessage(s, channelID, standings)
}

func (h *RefreshableBackend) RefreshMessage(s *discordgo.Session, event *discordgo.MessageReaction) error {
	standings, err := h.Leaderboard.GetStandings(context.TODO())
	if err != nil {
		return fmt.Errorf("error fetching leaderboard: %s", err)
	}
	embed, err := PrepareStandingsEmbed(&PrepareStandingsEmbedRequest{
		Standings:                   standings,
		AppendLastUpdatedEmbedField: true,
	})
	if err != nil {
		return fmt.Errorf("error preparing standings embed: %s", err)
	}
	_, err = s.ChannelMessageEditEmbed(event.ChannelID, event.MessageID, embed)
	if err != nil {
		return fmt.Errorf("error updating leaderboard message in channel %s: %s", event.ChannelID, err)
	}
	return nil
}
