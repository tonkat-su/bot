package emoji

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

func (p *Player) EmojiTextCode() string {
	return "<:" + p.EmojiName() + ":" + p.emojiID + ">"
}

func fillPlayerEmojis(input []*discordgo.Emoji, players []*Player, fill func(*Player) (string, error)) error {
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
			emojiId, err := fill(player)
			if err != nil {
				return err
			}
			player.emojiID = emojiId
		}
	}
	return nil
}

func fillEmoji(session *discordgo.Session, guildId string) func(*Player) (string, error) {
	return func(player *Player) (string, error) {
		face, err := mcuser.GetFace(player.Uuid)
		if err != nil {
			return "", fmt.Errorf("error getting face for %s: %s", player.Name, err.Error())
		}
		emojiParams := &discordgo.EmojiParams{
			Name:  player.EmojiName(),
			Image: dataurl.New(face, "image/png").String(),
		}
		emoji, err := session.GuildEmojiCreate(guildId, emojiParams, nil)
		if err != nil {
			return "", fmt.Errorf("error uploading emoji '%s': %s", player.EmojiName(), err.Error())
		}
		return emoji.ID, nil
	}
}

func SyncMinecraftAvatarsToEmoji(session *discordgo.Session, guildId string, players []*Player) error {
	guild, err := session.Guild(guildId)
	if err != nil {
		return err
	}

	return fillPlayerEmojis(guild.Emojis, players, fillEmoji(session, guildId))
}

func PlayerListEmojis(players []*Player) string {
	emojis := make([]string, len(players))
	for i, p := range players {
		emojis[i] = p.EmojiTextCode()
	}
	return strings.Join(emojis, " ")
}
