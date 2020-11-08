package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/tonkat-su/bot/users"
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

func mentionsUser(user *discordgo.User, users []*discordgo.User) bool {
	for _, v := range users {
		if v.ID == user.ID {
			return true
		}
	}
	return false
}

func reply(s *discordgo.Session, m *discordgo.MessageCreate, message string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", m.Author.Mention(), message))
	return err
}

func registerMinecraftGamer(svc *users.Service) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !mentionsUser(s.State.User, m.Mentions) {
			return
		}

		args := strings.Split(m.Content, " ")
		if len(args) < 2 || args[1] != "register" {
			return
		}

		const helptext = "register @discorduser <minecraft username or uuid>"
		if len(args) != 4 {
			err := reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}

		if len(m.Mentions) != 2 {
			err := reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}

		input := &users.RegisterInput{}
		if args[3] == "" {
			err := reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}
		_, err := uuid.Parse(args[3])
		if err != nil {
			input.MinecraftUsername = &args[3]
		} else {
			input.MinecraftUserId = &args[3]
		}

		var targetDiscordUser *discordgo.User
		for _, v := range m.Mentions {
			if v.ID == s.State.User.ID {
				continue
			}
			targetDiscordUser = v
		}
		input.DiscordUserId = targetDiscordUser.ID

		err = svc.Register(context.TODO(), input)
		if err != nil {
			log.Printf("error registering user: %s", err)
			sendErr := reply(s, m, "got an error registering user, try again later")
			if sendErr != nil {
				log.Printf("error sending message: %s", sendErr)
			}
			return
		}

		err = reply(s, m, "registration succeeded!")
		if err != nil {
			log.Printf("error sending message: %s", err)
		}
	}
}

func echo(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !mentionsUser(s.State.User, m.Mentions) || !strings.Contains(m.Content, "echo") {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "```\n"+m.Content+"\n```")
	if err != nil {
		log.Println(err)
	}
}
