package interactions

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/leaderboard"
)

func (srv *Server) leaderboard(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("error loading aws config: %s", err)
		writeResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	board, err := leaderboard.New(awsCfg, &leaderboard.Config{
		NamespacePrefix: srv.cfg.MinecraftServerName,
	})
	if err != nil {
		log.Printf("error instantiating leaderboard: %s", err)
		writeResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	standings, err := board.GetStandings(context.Background())
	if err != nil {
		log.Printf("error fetching leaderboard: %s", err)
		writeResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	messageEmbed, err := leaderboard.PrepareStandingsEmbed(&leaderboard.PrepareStandingsEmbedRequest{
		Standings: standings,
		Session:   s,
		GuildId:   srv.cfg.DiscordGuildId,
	})
	if err != nil {
		log.Printf("error preparing standings: %s", err)
		writeResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response := discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{messageEmbed},
		},
	}
	respondToInteraction(w, http.StatusOK, response)
}
