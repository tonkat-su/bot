package interactions

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/jltobler/go-rcon"
)

func (srv *Server) whitelist(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	log.Println("handling whitelist request")

	subcommand := event.ApplicationCommandData().Options[0].Options[0]
	rconClient := rcon.NewClient("rcon://"+srv.cfg.RconHostport, srv.cfg.RconPassword)

	var rconCommand string
	switch subcommand.Name {
	case "list":
		rconCommand = "whitelist list"
	case "add":
		for _, v := range subcommand.Options {
			if v.Name == "mc_user" {
				if username, ok := v.Value.(string); ok {
					rconCommand = fmt.Sprintf("whitelist add %s", username)
				}
			}
		}
	case "remove":
		for _, v := range subcommand.Options {
			if v.Name == "mc_user" {
				if username, ok := v.Value.(string); ok {
					rconCommand = fmt.Sprintf("whitelist remove %s", username)
				}
			}
		}
	default:
		log.Printf("invalid command: %s", subcommand.Name)
		writeResponse(w, http.StatusUnprocessableEntity, "invalid whitelist subcommand")
		return
	}

	log.Printf("sending rcon command: %s", rconCommand)
	output, err := rconClient.Send(rconCommand)
	if err != nil {
		log.Printf("error sending rcon command: %s", err.Error())
		writeResponse(w, http.StatusFailedDependency, err.Error())
	}

	err = replyToInteraction(event.ID, event.Token, output)
	if err != nil {
		log.Printf("error replying to interaction: %s", err.Error())
	}
	log.Println("rcon command successful")
	writeResponse(w, http.StatusOK, output)
}
