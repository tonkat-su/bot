package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func Reply(s *discordgo.Session, m *discordgo.MessageCreate, message string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", m.Author.Mention(), message))
	return err
}
