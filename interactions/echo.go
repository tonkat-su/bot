package interactions

import (
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (srv *Server) echo(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	content := event.ApplicationCommandData().Options[0].StringValue()
	log.Printf("echo %s", content)
	writeResponse(w, http.StatusOK, content)
}
