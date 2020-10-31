package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func anyGamers(cfg Config) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !(strings.HasPrefix(m.Content, "any gamers") || strings.HasPrefix(m.Content, "Any gamers")) {
			return
		}

		err := sendServerStatus(s, m, cfg, nil)
		if err != nil {
			log.Printf("error sending message: %s", err.Error())
			return
		}
	}
}
