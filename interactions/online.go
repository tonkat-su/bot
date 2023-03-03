package interactions

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/handlers/connected"
)

func (srv *Server) online(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	messageEmbed, err := connected.PrepareStatusEmbed(s, srv.cfg.DiscordGuildId, srv.cfg.MinecraftServerHostPort, srv.cfg.MinecraftServerName, srv.imgur)
	if err != nil {
		log.Printf("error rendering online message embed: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{messageEmbed},
		},
	})
	if err != nil {
		log.Printf("failed to encode body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
