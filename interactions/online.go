package interactions

import (
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/jltobler/go-rcon"
)

func (srv *Server) online(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	rconClient := rcon.NewClient("rcon://"+srv.cfg.RconHostport, srv.cfg.RconPassword)
	output, err := rconClient.Send("list")
	if err != nil {
		log.Printf("error sending list command: %s", err.Error())
	}

	writeResponse(w, http.StatusOK, output)
}
