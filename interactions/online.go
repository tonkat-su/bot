package interactions

import (
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/online"
)

func (srv *Server) online(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	prepareStatusResponse, err := online.PrepareStatus(&online.PrepareStatusRequest{
		Session:        s,
		GuildId:        srv.cfg.DiscordGuildId,
		ServerHostname: srv.cfg.MinecraftServerHost,
		ServerName:     srv.cfg.MinecraftServerName,
	})
	if err != nil {
		log.Printf("error rendering online message embed: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: prepareStatusResponse.MessageEmbeds,
		},
	}

	respondToInteraction(w, http.StatusOK, response)
}
