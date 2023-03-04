package leaderboard

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/emoji"
	"github.com/tonkat-su/bot/mcuser"
)

type PrepareStandingsEmbedRequest struct {
	Standings *Standings
	Session   *discordgo.Session
	GuildId   string
}

func PrepareStandingsEmbed(params *PrepareStandingsEmbedRequest) (*discordgo.MessageEmbed, error) {
	players := make([]*emoji.Player, len(params.Standings.SortedStandings))
	for i, v := range params.Standings.SortedStandings {
		username, err := mcuser.GetUsername(v.PlayerId)
		if err != nil {
			return nil, fmt.Errorf("error getting username: %s", err)
		}
		players[i] = &emoji.Player{
			Name: username,
			Uuid: v.PlayerId,
		}
	}

	err := emoji.SyncMinecraftAvatarsToEmoji(params.Session, params.GuildId, players)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	for i, v := range params.Standings.SortedStandings {
		fmt.Fprintf(&builder, "%s %s: %d", players[i].EmojiTextCode(), players[i].Name, v.Score)
		if i != len(params.Standings.SortedStandings)-1 {
			builder.WriteString("\n")
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: `biggest nerds on the server
(in the last 7 days)`,
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: builder.String(),
			},
		},
	}

	return embed, nil
}
