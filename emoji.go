package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/mcuser"
)

type Player struct {
	Name string
	Uuid string
}

func (p *Player) EmojiName() string {
	return strings.ToLower(p.Name) + "Face"
}

func filterPlayersThatNeedEmoji(input []*discordgo.Emoji, players []Player) []Player {
	e := make(map[string]*discordgo.Emoji)
	for _, emoji := range input {
		if strings.HasSuffix(emoji.Name, "Face") {
			e[emoji.Name] = emoji
		}
	}

	var index int
	for _, player := range players {
		if _, ok := e[player.EmojiName()]; !ok {
			players[index] = player
			index++
		}
	}

	return players[:index]
}

type playerEmoji struct {
	code  string
	image string
}

func fetchAvatarsAndPrepareEmoji(players []Player) ([]playerEmoji, error) {
	emoji := make([]playerEmoji, len(players))
	for i, player := range players {
		face, err := mcuser.GetFace(player.Uuid)
		if err != nil {
			return nil, fmt.Errorf("error getting face for %s: %s", player.Name, err.Error())
		}
		emoji[i] = playerEmoji{
			code:  player.EmojiName(),
			image: base64.StdEncoding.EncodeToString(face),
		}
	}
	return emoji, nil
}

func syncMinecraftAvatarsToEmoji(session *discordgo.Session, guildId string, players []Player) error {
	guild, err := session.Guild(guildId)
	if err != nil {
		return err
	}
	emoji, err := fetchAvatarsAndPrepareEmoji(filterPlayersThatNeedEmoji(guild.Emojis, players))
	if err != nil {
		return err
	}
	for _, e := range emoji {
		_, err = session.GuildEmojiCreate(guildId, e.code, e.image, nil)
		if err != nil {
			return fmt.Errorf("error uploading emoji '%s': %s", e.code, err.Error())
		}
	}
	return nil
}

func playerListEmojis(players []Player) string {
	emojis := make([]string, len(players))
	for i, p := range players {
		emojis[i] = ":" + p.EmojiName() + ":"
	}
	return strings.Join(emojis, " ")
}
