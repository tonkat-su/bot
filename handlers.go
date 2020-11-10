package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/tonkat-su/bot/leaderboard"
	"github.com/tonkat-su/bot/mcuser"
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

func lookupUser(usersService *users.Service) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !mentionsUser(s.State.User, m.Mentions) {
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
			if sendErr := reply(s, m, helptext); sendErr != nil {
				log.Printf("error sending reply: %s", sendErr)
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
				if sendErr := reply(s, m, "got an error looking up user, try again later"); sendErr != nil {
					log.Printf("error sending reply: %s", sendErr)
				}
				return
			}
		} else {
			var err error
			storedUserInfo, err = usersService.LookupByMinecraftUsername(context.TODO(), &users.LookupInput{Id: args[2]})
			if err != nil {
				log.Printf("error looking up user: %s", err)
				if sendErr := reply(s, m, "got an error looking up user, try again later"); sendErr != nil {
					log.Printf("error sending reply: %s", sendErr)
				}
				return
			}
		}

		if storedUserInfo == nil {
			if sendErr := reply(s, m, "user not found"); sendErr != nil {
				log.Printf("error sending reply: %s", sendErr)
			}
			return
		}

		discordUser, err := s.User(storedUserInfo.DiscordUserId)
		if err != nil {
			log.Printf("error looking up discord username by id: %s", err)
			if sendErr := reply(s, m, "error looking up discord username"); sendErr != nil {
				log.Printf("error sending reply: %s", sendErr)
			}
			return
		}

		msg := &discordgo.MessageEmbed{
			Title: "user registration result",
			Color: 0x43b581,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "discord user",
					Value: discordUser.Username,
				},
				{
					Name:  "minecraft name",
					Value: storedUserInfo.MinecraftUsername,
				},
				{
					Name:  "minecraft id",
					Value: storedUserInfo.MinecraftUserId,
				},
			},
		}
		if _, sendErr := s.ChannelMessageSendEmbed(m.ChannelID, msg); sendErr != nil {
			log.Printf("error sending reply: %s", sendErr)
		}
	}
}

func leaderboardRequestHandler(lboard *leaderboard.Service) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !mentionsUser(s.State.User, m.Mentions) {
			return
		}

		args := strings.Split(m.Content, " ")
		if len(args) < 2 {
			return
		}
		if args[1] != "leaderboard" {
			return
		}

		standings, err := lboard.GetStandings(context.TODO())
		if err != nil {
			log.Printf("error getting standings: %s", err)
			if sendErr := reply(s, m, "error fetching standings"); sendErr != nil {
				log.Printf("error replying: %s", err)
			}
			return
		}
		if standings == nil {
			return
		}

		_, err = sendStandingsMessage(s, m.ChannelID, standings)
		if err != nil {
			log.Println(err)
		}
	}
}

func prepareStandingsEmbed(standings *leaderboard.Standings) (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title: `biggest nerds on the server
(in the last 7 days)`,
		Fields: make([]*discordgo.MessageEmbedField, len(standings.SortedStandings)),
	}
	for i, v := range standings.SortedStandings {
		username, err := mcuser.GetUsername(v.PlayerId)
		if err != nil {
			return nil, fmt.Errorf("error getting username: %s", err)
		}
		embed.Fields[i] = &discordgo.MessageEmbedField{
			Name:  username,
			Value: fmt.Sprintf("%d cat treats", v.Score),
		}
	}
	return embed, nil
}

func sendStandingsMessage(s *discordgo.Session, channelID string, standings *leaderboard.Standings) (*discordgo.Message, error) {
	embed, err := prepareStandingsEmbed(standings)
	if err != nil {
		return nil, err
	}
	return s.ChannelMessageSendEmbed(channelID, embed)
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

type leaderboardHandlerService struct {
	sync.RWMutex

	leaderboard *leaderboard.Service

	// guild id to pinnedMessageId
	messages map[string]string
}

func (svc *leaderboardHandlerService) onConnect(s *discordgo.Session, event *discordgo.Ready) {
	for _, guild := range event.Guilds {
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			log.Fatalf("unable to fetch guild channels for leaderboard state: %s", err)
		}

		var boardChannel string
		for _, channel := range channels {
			if channel.Name == "leaderboard" {
				boardChannel = channel.ID
				log.Printf("found leaderboard guild:%s channel:%s", guild.ID, channel.ID)
				pinnedMessages, err := s.ChannelMessagesPinned(channel.ID)
				if err != nil {
					log.Fatalf("unable to fetch pinned messages for channel '%s': %s", channel.ID, err)
				}
				for _, message := range pinnedMessages {
					if message.Author.ID == s.State.User.ID {
						log.Printf("found leaderboard channel:%s message:%s", channel.ID, message.ID)
						svc.Lock()
						svc.messages[guild.ID] = message.ID
						svc.Unlock()
						break
					}
				}
				break
			}
		}

		if boardChannel == "" {
			log.Printf("creating leaderboard channel for guild:%s", guild.ID)
			channel, err := s.GuildChannelCreate(guild.ID, "leaderboard", discordgo.ChannelTypeGuildText)
			if err != nil {
				log.Printf("unable to create leaderboard for guild (%s:%s): %s", guild.Name, guild.ID, err)
				continue
			}
			boardChannel = channel.ID
		}

		if _, ok := svc.messages[guild.ID]; !ok {
			log.Printf("creating leaderboard message for channel:%s", boardChannel)
			standings, err := svc.leaderboard.GetStandings(context.TODO())
			if err != nil {
				log.Printf("error fetching leaderboard: %s", err)
			}
			msg, err := sendStandingsMessage(s, boardChannel, standings)
			if err != nil {
				log.Printf("error sending leaderboard message to channel (%s): %s", boardChannel, err)
			}
			err = s.ChannelMessagePin(boardChannel, msg.ID)
			if err != nil {
				log.Printf("error pinning leaderboard message (%s) to channel (%s): %s", msg.ID, boardChannel, err)
			}
			svc.Lock()
			svc.messages[guild.ID] = msg.ID
			svc.Unlock()
		}

		err = s.MessageReactionsRemoveAll(boardChannel, svc.messages[guild.ID])
		if err != nil {
			log.Printf("error removing all reactions from leaderboard message in channel %s: %s", boardChannel, err)
		}

		err = s.MessageReactionAdd(boardChannel, svc.messages[guild.ID], "♻️")
		if err != nil {
			log.Printf("error adding refresh reaction to leaderboard message in channel %s: %s", svc.messages[guild.ID], err)
		}
	}
}

func (svc *leaderboardHandlerService) updateLeaderboard(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.MessageReaction.UserID == s.State.User.ID || svc.messages[m.MessageReaction.GuildID] != m.MessageReaction.MessageID {
		return
	}

	err := s.MessageReactionRemove(m.MessageReaction.ChannelID, m.MessageReaction.MessageID, `♻️`, m.MessageReaction.UserID)
	if err != nil {
		log.Printf("error removing reaction from leaderboard message: %s", err)
		return
	}

	standings, err := svc.leaderboard.GetStandings(context.TODO())
	if err != nil {
		log.Printf("error fetching leaderboard: %s", err)
		return
	}
	embed, err := prepareStandingsEmbed(standings)
	if err != nil {
		log.Printf("error preparing standings embed: %s", err)
		return
	}
	_, err = s.ChannelMessageEditEmbed(m.MessageReaction.ChannelID, m.MessageReaction.MessageID, embed)
	if err != nil {
		log.Printf("error updating leaderboard message in channel %s: %s", m.MessageReaction.ChannelID, err)
	}
}
