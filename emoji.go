package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/mcuser"
	"github.com/vincent-petithory/dataurl"
)

type Player struct {
	Name    string
	Uuid    string
	emojiID string
}

func (p *Player) EmojiName() string {
	return strings.ToLower(p.Name) + "Face"
}

func fillPlayerEmojis(input []*discordgo.Emoji, players []*Player, fill func(*Player) error) error {
	e := make(map[string]*discordgo.Emoji)
	for _, emoji := range input {
		if strings.HasSuffix(emoji.Name, "Face") {
			e[emoji.Name] = emoji
		}
	}

	for _, player := range players {
		emoji, ok := e[player.EmojiName()]
		if ok {
			player.emojiID = emoji.ID
		} else {
			err := fill(player)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func fillEmoji(session *discordgo.Session, guildId string) func(*Player) error {
	return func(player *Player) error {
		face, err := mcuser.GetFace(player.Uuid)
		if err != nil {
			return fmt.Errorf("error getting face for %s: %s", player.Name, err.Error())
		}
		emoji, err := session.GuildEmojiCreate(guildId, player.EmojiName(), dataurl.New(face, "image/png").String(), nil)
		if err != nil {
			return fmt.Errorf("error uploading emoji '%s': %s", player.EmojiName(), err.Error())
		}
		player.emojiID = emoji.ID
		return nil
	}
}

func syncMinecraftAvatarsToEmoji(session *discordgo.Session, guildId string, players []*Player) error {
	guild, err := session.Guild(guildId)
	if err != nil {
		return err
	}

	return fillPlayerEmojis(guild.Emojis, players, fillEmoji(session, guildId))
}

func playerListEmojis(players []*Player) string {
	emojis := make([]string, len(players))
	for i, p := range players {
		emojis[i] = "<:" + p.EmojiName() + ":" + p.emojiID + ">"
	}
	return strings.Join(emojis, " ")
}
