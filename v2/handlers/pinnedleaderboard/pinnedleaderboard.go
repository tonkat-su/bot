package pinnedleaderboard

import (
	"context"
	"fmt"

	"github.com/andersfylling/disgord"
	"github.com/tonkat-su/bot/v2/handlers"
	"github.com/tonkat-su/bot/v2/leaderboard"
	"github.com/tonkat-su/bot/v2/mcuser"
)

func prepareStandingsEmbed(standings *leaderboard.Standings) (*disgord.Embed, error) {
	embed := &disgord.Embed{
		Title: `biggest nerds on the server
(in the last 7 days)`,
		Fields: make([]*disgord.EmbedField, len(standings.SortedStandings), len(standings.SortedStandings)+1),
	}
	for i, v := range standings.SortedStandings {
		username, err := mcuser.GetUsername(v.PlayerId)
		if err != nil {
			return nil, fmt.Errorf("error getting username: %s", err)
		}
		embed.Fields[i] = &disgord.EmbedField{
			Name:  username,
			Value: fmt.Sprintf("%d cat treats", v.Score),
		}
	}

	var err error
	embed.Fields, err = handlers.AppendLastUpdatedEmbedField(embed.Fields, standings.LastUpdated)
	if err != nil {
		return nil, err
	}

	return embed, nil
}

func sendStandingsMessage(session disgord.Session, channelID disgord.Snowflake, standings *leaderboard.Standings) (*disgord.Message, error) {
	embed, err := prepareStandingsEmbed(standings)
	if err != nil {
		return nil, err
	}
	return session.SendMsg(channelID, embed)
}

type RefreshableBackend struct {
	Leaderboard *leaderboard.Service
}

func (h *RefreshableBackend) CreateRefreshableMessage(session disgord.Session, guildID, channelID disgord.Snowflake) (*disgord.Message, error) {
	standings, err := h.Leaderboard.GetStandings(context.TODO())
	if err != nil {
		return nil, err
	}
	return sendStandingsMessage(session, channelID, standings)
}

func (h *RefreshableBackend) RefreshMessage(session disgord.Session, event *disgord.MessageReactionAdd) error {
	standings, err := h.Leaderboard.GetStandings(context.TODO())
	if err != nil {
		return fmt.Errorf("error fetching leaderboard: %s", err)
	}
	embed, err := prepareStandingsEmbed(standings)
	if err != nil {
		return fmt.Errorf("error preparing standings embed: %s", err)
	}
	_, err = session.Channel(event.ChannelID).Message(event.MessageID).SetEmbed(embed)
	if err != nil {
		return fmt.Errorf("error updating leaderboard message in channel %s: %s", event.ChannelID, err)
	}
	return nil
}
