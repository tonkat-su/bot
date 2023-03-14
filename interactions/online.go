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

	/*
		can't use respondToInteraction because we need to use multipart response
		to upload files
	*/
	contentType, responseBody, err := discordgo.MultipartBodyWithJSON(response, prepareStatusResponse.Files)
	if err != nil {
		log.Printf("error preparing multipart body: %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseBody)
	if err != nil {
		log.Printf("error writing response body: %s", err.Error())
		return
	}
}
