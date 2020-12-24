package refreshable

import (
	"context"
	"log"
	"sync"

	"github.com/andersfylling/disgord"
)

type Backend interface {
	CreateRefreshableMessage(s disgord.Session, guildID disgord.Snowflake, channelID disgord.Snowflake) (*disgord.Message, error)
	RefreshMessage(disgord.Session, *disgord.MessageReactionAdd) error
}

type Handler struct {
	sync.Mutex

	Backend           Backend
	PinnedChannelName string

	// guild id to pinned message
	messages map[disgord.Snowflake]*disgord.Message
}

func (h *Handler) AddHandlers(s disgord.Session) {
	s.Gateway().Ready(h.OnReady)
	s.Gateway().MessageReactionAdd(h.OnMessageReactionAdd)
}

func (h *Handler) OnReady(s disgord.Session, event *disgord.Ready) {
	h.Lock()
	h.messages = map[disgord.Snowflake]*disgord.Message{}
	h.Unlock()

	currentUser, err := s.CurrentUser().Get()
	if err != nil {
		log.Fatalf("unable to get current user: %s", err)
	}

	for _, guild := range event.Guilds {
		channels, err := s.Guild(guild.ID).GetChannels()
		if err != nil {
			log.Fatalf("unable to fetch guild channels for refreshable state: %s", err)
		}

		var boardChannel disgord.Snowflake
		for _, channel := range channels {
			if channel.Name == h.PinnedChannelName {
				boardChannel = channel.ID
				log.Printf("found refreshable %s guild:%s channel:%s", h.PinnedChannelName, guild.ID, channel.ID)
				pinnedMessages, err := s.Channel(channel.ID).GetPinnedMessages()
				if err != nil {
					log.Fatalf("unable to fetch pinned messages for %s channel '%s': %s", h.PinnedChannelName, channel.ID, err)
				}
				for _, message := range pinnedMessages {
					if message.Author.ID == currentUser.ID {
						log.Printf("found %s channel:%s message:%s", h.PinnedChannelName, channel.ID, message.ID)
						h.Lock()
						h.messages[guild.ID] = message
						h.Unlock()
						break
					}
				}
				break
			}
		}

		if boardChannel.IsZero() {
			log.Printf("creating refreshable %s channel for guild:%s", h.PinnedChannelName, guild.ID)
			channel, err := s.Guild(guild.ID).CreateChannel(h.PinnedChannelName, &disgord.CreateGuildChannelParams{
				Name: h.PinnedChannelName,
			})
			if err != nil {
				log.Printf("unable to create pinned %s channel for guild (%s): %s", h.PinnedChannelName, guild.ID, err)
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
			err = s.Channel(boardChannel).Message(msg.ID).Pin()
			if err != nil {
				log.Printf("error pinning message (%s) to %s channel (%s): %s", h.PinnedChannelName, msg.ID, boardChannel, err)
			}
			h.Lock()
			h.messages[guild.ID] = msg
			h.Unlock()
		}

		ctx := context.Background()
		err = removeAllReactions(ctx, s, h.messages[guild.ID])
		if err != nil {
			log.Printf("error removing all reactions from pinned message in %s channel %s: %s", h.PinnedChannelName, boardChannel, err)
		}

		err = h.messages[guild.ID].React(ctx, s, "♻️")
		if err != nil {
			log.Printf("error adding refresh reaction to pinned message in %s channel %s: %s", h.PinnedChannelName, h.messages[guild.ID].ChannelID, err)
		}
	}
}

func removeAllReactions(ctx context.Context, s disgord.Session, msg *disgord.Message) error {
	for _, reaction := range msg.Reactions {
		err := s.Channel(msg.ChannelID).Message(msg.ID).Reaction(reaction.Emoji).DeleteUser(reaction.Emoji.User.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) OnMessageReactionAdd(s disgord.Session, event *disgord.MessageReactionAdd) {
	currentUser, err := s.CurrentUser().Get()
	if err != nil {
		log.Printf("error getting current user: %s", err)
		return
	}

	channel, err := s.Channel(event.ChannelID).Get()
	if err != nil {
		log.Printf("error getting channel for message reaction: %s", err)
		return
	}

	cachedMessage := h.messages[channel.GuildID]
	if event.UserID == currentUser.ID || cachedMessage.ID != event.MessageID {
		return
	}

	err = s.Channel(event.ChannelID).Message(event.MessageID).Reaction(`♻️`).DeleteUser(event.UserID)
	if err != nil {
		log.Printf("error removing reaction from refreshable message %s (%s): %s", h.PinnedChannelName, event.MessageID, err)
		return
	}

	err = h.Backend.RefreshMessage(s, event)
	if err != nil {
		log.Println(err)
	}
}
