package echo

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/v2/handlers"
)

func Echo(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !handlers.MentionsUser(s.State.User, m.Mentions) || !strings.Contains(m.Content, "echo") {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "```\n"+m.Content+"\n```")
	if err != nil {
		log.Println(err)
	}
}
