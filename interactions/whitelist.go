package interactions

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jltobler/go-rcon"
	"github.com/tonkat-su/bot/emoji"
)

func (srv *Server) whitelist(w http.ResponseWriter, event discordgo.Interaction, s *discordgo.Session) {
	log.Println("handling whitelist request")

	subcommand := event.ApplicationCommandData().Options[0]
	rconClient := rcon.NewClient("rcon://"+srv.cfg.RconHostport, srv.cfg.RconPassword)

	var rconCommand string
	switch subcommand.Name {
	case "list":
		rconCommand = "whitelist list"
	case "add":
		for _, v := range subcommand.Options {
			if v.Name == "username" {
				if username, ok := v.Value.(string); ok {
					rconCommand = fmt.Sprintf("whitelist add %s", username)
				}
			}
		}
	case "remove":
		for _, v := range subcommand.Options {
			if v.Name == "username" {
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

	if subcommand.Name == "list" {
		embed, err := prepareWhitelistedEmbed(&prepareWhitelistedEmbedParams{
			Session:        s,
			DiscordGuildId: srv.cfg.DiscordGuildId,
			Players:        strings.Split(strings.Split(output, ": ")[1], ", "),
		})
		if err != nil {
			log.Printf("error syncing avatars to emoji for whitelist list: %s", err.Error())
			writeResponse(w, http.StatusFailedDependency, err.Error())
		}
		respondToInteraction(w, http.StatusOK, discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		return
	}

	writeResponse(w, http.StatusOK, output)
	log.Println("rcon command successful")
}

type prepareWhitelistedEmbedParams struct {
	Session        *discordgo.Session
	Players        []string
	DiscordGuildId string
}

func prepareWhitelistedEmbed(params *prepareWhitelistedEmbedParams) (*discordgo.MessageEmbed, error) {
	players := []*emoji.Player{}
	for _, name := range params.Players {
		players = append(players, &emoji.Player{
			Name: name,
		})
	}
	err := emoji.SyncMinecraftAvatarsToEmoji(params.Session, params.DiscordGuildId, players)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	for i, v := range players {
		fmt.Fprintf(&builder, "%s %s", v.EmojiTextCode(), v.Name)
		if i != len(players)-1 {
			builder.WriteString("\n")
		}
	}

	return &discordgo.MessageEmbed{
		Title: "Our Froggy Friends",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: builder.String(),
			},
		},
	}, nil
}
