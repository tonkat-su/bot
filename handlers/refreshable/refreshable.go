package refreshable

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Backend interface {
	CreateRefreshableMessage(s *discordgo.Session, guildID string, channelID string) (*discordgo.Message, error)
	RefreshMessage(*discordgo.Session, *discordgo.MessageReaction) error
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

func (h *Handler) OnConnect(s *discordgo.Session, event *discordgo.Ready) {
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
				log.Printf("found refreshable %s guild:%s channel:%s", h.PinnedChannelName, guild.ID, channel.ID)
				pinnedMessages, err := s.ChannelMessagesPinned(channel.ID)
				if err != nil {
					log.Fatalf("unable to fetch pinned messages for %s channel '%s': %s", h.PinnedChannelName, channel.ID, err)
				}
				for _, message := range pinnedMessages {
					if message.Author.ID == s.State.User.ID {
						log.Printf("found %s channel:%s message:%s", h.PinnedChannelName, channel.ID, message.ID)
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
			log.Printf("creating refreshable %s channel for guild:%s", h.PinnedChannelName, guild.ID)
			channel, err := s.GuildChannelCreate(guild.ID, h.PinnedChannelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				log.Printf("unable to create pinned %s channel for guild (%s:%s): %s", h.PinnedChannelName, guild.Name, guild.ID, err)
				continue
			}
			boardChannel = channel.ID
		}

		if _, ok := h.messages[guild.ID]; !ok {
			msg, err := h.Backend.CreateRefreshableMessage(s, guild.ID, boardChannel)
			if err != nil {
				log.Printf("error sending message to %s channel (%s): %s", h.PinnedChannelName, boardChannel, err)
				return
			}
			err = s.ChannelMessagePin(boardChannel, msg.ID)
			if err != nil {
				log.Printf("error pinning message (%s) to %s channel (%s): %s", h.PinnedChannelName, msg.ID, boardChannel, err)
			}
			h.Lock()
			h.messages[guild.ID] = msg.ID
			h.Unlock()
		}

		err = s.MessageReactionsRemoveAll(boardChannel, h.messages[guild.ID])
		if err != nil {
			log.Printf("error removing all reactions from pinned message in %s channel %s: %s", h.PinnedChannelName, boardChannel, err)
		}

		err = s.MessageReactionAdd(boardChannel, h.messages[guild.ID], "♻️")
		if err != nil {
			log.Printf("error adding refresh reaction to pinned message in %s channel %s: %s", h.PinnedChannelName, h.messages[guild.ID], err)
		}
	}
}

func (h *Handler) OnMessageReactionAdd(s *discordgo.Session, event *discordgo.MessageReactionAdd) {
	if event.MessageReaction.UserID == s.State.User.ID || h.messages[event.MessageReaction.GuildID] != event.MessageReaction.MessageID {
		return
	}

	err := s.MessageReactionRemove(event.MessageReaction.ChannelID, event.MessageReaction.MessageID, `♻️`, event.MessageReaction.UserID)
	if err != nil {
		log.Printf("error removing reaction from refreshable message %s (%s): %s", h.PinnedChannelName, event.MessageReaction.MessageID, err)
		return
	}

	err = h.Backend.RefreshMessage(s, event.MessageReaction)
	if err != nil {
		log.Println(err)
	}
}
