package interactions

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (srv *Server) echo(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	writeResponse(w, http.StatusOK, event.Message.Content)
}
