package interactions

import (
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/online"
)

func (srv *Server) online(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	messageEmbed, err := online.PrepareStatusEmbed(&online.PrepareStatusEmbedRequest{
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

	w.WriteHeader(http.StatusOK)
}
