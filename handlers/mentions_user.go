package handlers

import "github.com/bwmarrin/discordgo"

func MentionsUser(user *discordgo.User, users []*discordgo.User) bool {
	for _, v := range users {
		if v.ID == user.ID {
			return true
		}
	}
	return false
}
