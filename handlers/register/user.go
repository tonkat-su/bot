package register

import (
	"context"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/tonkat-su/bot/handlers"
	"github.com/tonkat-su/bot/mcuser"
	"github.com/tonkat-su/bot/users"
)

func RegisterMinecraftGamer(svc *users.Service) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !handlers.MentionsUser(s.State.User, m.Mentions) {
			return
		}

		args := strings.Split(m.Content, " ")
		if len(args) < 2 || args[1] != "register" {
			return
		}

		const helptext = "register @discorduser <minecraft username or uuid>"
		if len(args) != 4 {
			err := handlers.Reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}

		if len(m.Mentions) != 2 {
			err := handlers.Reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}

		input := &users.RegisterInput{}
		if args[3] == "" {
			err := handlers.Reply(s, m, helptext)
			if err != nil {
				log.Printf("error sending message: %s", err)
			}
			return
		}
		_, err := uuid.Parse(args[3])
		if err != nil {
			input.MinecraftUserId, err = mcuser.GetUuid(args[3])
			if err != nil {
				log.Printf("error looking up minecraft uuid: %s", err)
				sendErr := handlers.Reply(s, m, "error looking up minecraft uuid")
				if sendErr != nil {
					log.Printf("error sending message: %s", err)
				}
				return
			}
		} else {
			input.MinecraftUserId = args[3]
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
			sendErr := handlers.Reply(s, m, "got an error registering user, try again later")
			if sendErr != nil {
				log.Printf("error sending message: %s", sendErr)
			}
			return
		}

		err = handlers.Reply(s, m, "registration succeeded!")
		if err != nil {
			log.Printf("error sending message: %s", err)
		}
	}
}

func LookupUser(usersService *users.Service) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !handlers.MentionsUser(s.State.User, m.Mentions) {
			return
		}

		args := strings.Split(m.Content, " ")
		if len(args) < 2 {
			return
		}
		if args[1] != "lookup" {
			return
		}

		const helptext = "lookup <@discordUser or minecraftUsername>"
		if len(args) == 2 || len(args) > 3 {
			if sendErr := handlers.Reply(s, m, helptext); sendErr != nil {
				log.Printf("error sending handlers.Reply: %s", sendErr)
			}
			return
		}

		var storedUserInfo *users.LookupOutput
		if len(m.Mentions) == 2 {
			var targetDiscordUser *discordgo.User
			for _, v := range m.Mentions {
				if v.ID == s.State.User.ID {
					continue
				}
				targetDiscordUser = v
			}
			var err error
			storedUserInfo, err = usersService.LookupByDiscordId(context.TODO(), &users.LookupInput{Id: targetDiscordUser.ID})
			if err != nil {
				log.Printf("error looking up user: %s", err)
				if sendErr := handlers.Reply(s, m, "got an error looking up user, try again later"); sendErr != nil {
					log.Printf("error sending handlers.Reply: %s", sendErr)
				}
				return
			}
		} else {
			minecraftUuid, err := mcuser.GetUuid(args[2])
			if err != nil {
				log.Printf("error getting minecraft uuid: %s", err)
				if sendErr := handlers.Reply(s, m, "got an error looking up user, try again later"); sendErr != nil {
					log.Printf("error sending handlers.Reply: %s", sendErr)
				}
				return
			}
			storedUserInfo, err = usersService.LookupByMinecraftId(context.TODO(), &users.LookupInput{Id: minecraftUuid})
			if err != nil {
				log.Printf("error looking up user: %s", err)
				if sendErr := handlers.Reply(s, m, "got an error looking up user, try again later"); sendErr != nil {
					log.Printf("error sending handlers.Reply: %s", sendErr)
				}
				return
			}
		}

		if storedUserInfo == nil {
			if sendErr := handlers.Reply(s, m, "user not found"); sendErr != nil {
				log.Printf("error sending handlers.Reply: %s", sendErr)
			}
			return
		}

		discordUser, err := s.User(storedUserInfo.DiscordUserId)
		if err != nil {
			log.Printf("error looking up discord username by id: %s", err)
			if sendErr := handlers.Reply(s, m, "error looking up discord username"); sendErr != nil {
				log.Printf("error sending handlers.Reply: %s", sendErr)
			}
			return
		}

		minecraftUsername, err := mcuser.GetUsername(storedUserInfo.MinecraftUserId)
		if err != nil {
			log.Printf("error looking up minecraft username: %s", err)
			if sendErr := handlers.Reply(s, m, "error looking up minecraft username"); sendErr != nil {
				log.Printf("error sending handlers.Reply: %s", sendErr)
			}
			return
		}

		msg := &discordgo.MessageEmbed{
			Title: "user registration result",
			Color: 0x43b581,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "discord username",
					Value: discordUser.Username,
				},
				{
					Name:  "discord id",
					Value: storedUserInfo.DiscordUserId,
				},
				{
					Name:  "minecraft username",
					Value: minecraftUsername,
				},
				{
					Name:  "minecraft id",
					Value: storedUserInfo.MinecraftUserId,
				},
			},
		}
		if _, sendErr := s.ChannelMessageSendEmbed(m.ChannelID, msg); sendErr != nil {
			log.Printf("error sending handlers.Reply: %s", sendErr)
		}
	}
}
