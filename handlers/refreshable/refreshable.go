package refreshable

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Backend interface {
	CreateRefreshableMessage(s *discordgo.Session, guildID string, channelID string) (*discordgo.Message, error)
	RefreshMessage(s *discordgo.Session, channelID, messageID string) error
}

type Handler struct {
	sync.Mutex

	Backend           Backend
	PinnedChannelName string

	// guild id to pinnedMessageId
	messages map[string]string
}

func (h *Handler) AddHandlers(s *discordgo.Session) {
	s.AddHandler(h.OnConnect)
	s.AddHandler(h.OnMessageReactionAdd)
}

func (h *Handler) OnConnect() func(*discordgo.Session, *discordgo.Ready) {
	return func(s *discordgo.Session, event *discordgo.Ready) {
		h.Lock()
		h.messages = map[string]string{}
		h.Unlock()

		for _, guild := range event.Guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				log.Fatalf("unable to fetch guild channels for refreshable state: %s", err)
			}

			var boardChannel string
			for _, channel := range channels {
				if channel.Name == h.PinnedChannelName {
					boardChannel = channel.ID
					log.Printf("found refreshable guild:%s channel:%s", guild.ID, channel.ID)
					pinnedMessages, err := s.ChannelMessagesPinned(channel.ID)
					if err != nil {
						log.Fatalf("unable to fetch pinned messages for channel '%s': %s", channel.ID, err)
					}
					for _, message := range pinnedMessages {
						if message.Author.ID == s.State.User.ID {
							log.Printf("found channel:%s message:%s", channel.ID, message.ID)
							h.Lock()
							h.messages[guild.ID] = message.ID
							h.Unlock()
							break
						}
					}
					break
				}
			}

			if boardChannel == "" {
				log.Printf("creating refreshable channel for guild:%s", guild.ID)
				channel, err := s.GuildChannelCreate(guild.ID, h.PinnedChannelName, discordgo.ChannelTypeGuildText)
				if err != nil {
					log.Printf("unable to create pinned channel for guild (%s:%s): %s", guild.Name, guild.ID, err)
					continue
				}
				boardChannel = channel.ID
			}

			if _, ok := h.messages[guild.ID]; !ok {
				msg, err := h.Backend.CreateRefreshableMessage(s, guild.ID, boardChannel)
				if err != nil {
					log.Printf("error sending message to channel (%s): %s", boardChannel, err)
					return
				}
				err = s.ChannelMessagePin(boardChannel, msg.ID)
				if err != nil {
					log.Printf("error pinning message (%s) to channel (%s): %s", msg.ID, boardChannel, err)
				}
				h.Lock()
				h.messages[guild.ID] = msg.ID
				h.Unlock()
			}

			err = s.MessageReactionsRemoveAll(boardChannel, h.messages[guild.ID])
			if err != nil {
				log.Printf("error removing all reactions from pinned message in channel %s: %s", boardChannel, err)
			}

			err = s.MessageReactionAdd(boardChannel, h.messages[guild.ID], "♻️")
			if err != nil {
				log.Printf("error adding refresh reaction to pinned message in channel %s: %s", h.messages[guild.ID], err)
			}
		}

	}
}

func (h *Handler) OnMessageReactionAdd() func(*discordgo.Session, *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, event *discordgo.MessageReactionAdd) {
		if event.MessageReaction.UserID == s.State.User.ID || h.messages[event.MessageReaction.GuildID] != event.MessageReaction.MessageID {
			return
		}

		err := s.MessageReactionRemove(event.MessageReaction.ChannelID, event.MessageReaction.MessageID, `♻️`, event.MessageReaction.UserID)
		if err != nil {
			log.Printf("error removing reaction from refreshable message (%s): %s", event.MessageReaction.MessageID, err)
			return
		}

		err = h.Backend.RefreshMessage(s, event.MessageReaction.ChannelID, event.MessageReaction.MessageID)
		if err != nil {
			log.Println(err)
		}
	}
}
