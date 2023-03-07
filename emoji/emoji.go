package emoji

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/mcuser"
	"github.com/vincent-petithory/dataurl"
)

type Player struct {
	Name string
	Uuid string

	emojiId string
}

func (p *Player) EmojiName() string {
	return strings.ToLower(p.Name) + "Face"
}

func (p *Player) EmojiTextCode() string {
	return "<:" + p.EmojiName() + ":" + p.emojiId + ">"
}

func fillPlayerEmojis(input []*discordgo.Emoji, players []*Player, fill func(*Player) error) error {
	// TODO: cache this
	e := make(map[string]*discordgo.Emoji)
	for _, emoji := range input {
		if strings.HasSuffix(emoji.Name, "Face") {
			e[emoji.Name] = emoji
		}
	}

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *Player) {
			if e, ok := e[player.EmojiName()]; ok {
				player.emojiId = e.ID
			}
			err := fill(player)
			if err != nil {
				log.Printf("error filling player emoji: %s", err.Error())
			}
			wg.Done()
		}(player)
	}

	wg.Wait()
	return nil
}

func fillEmoji(session *discordgo.Session, guildId string) func(*Player) error {
	return func(player *Player) error {
		face, err := mcuser.GetFace(player.Name)
		if err != nil {
			return fmt.Errorf("error getting face for %s: %s", player.Name, err.Error())
		}

		if !checkIfEmojiNeedsUpdate(player.emojiId, face) {
			return nil
		}

		// only delete if an emoji currently exists
		if player.emojiId != "" {
			log.Printf("deleting old emoji for player %s, id %s", player.Name, player.emojiId)
			// delete existing emoji
			err = session.GuildEmojiDelete(guildId, player.emojiId)
			if err != nil {
				// eat the error and just create on top of it
				log.Printf("error deleting existing emoji id %s: %s", player.emojiId, err.Error())
			}
		}

		emojiParams := &discordgo.EmojiParams{
			Name:  player.EmojiName(),
			Image: dataurl.New(face, "image/png").String(),
		}
		emoji, err := session.GuildEmojiCreate(guildId, emojiParams)
		if err != nil {
			return fmt.Errorf("error uploading emoji '%s': %s", player.EmojiName(), err.Error())
		}
		log.Printf("created emoji for player %s, id %s", player.Name, emoji.ID)
		player.emojiId = emoji.ID
		return nil
	}
}

func HydrateEmojiIds(session *discordgo.Session, guildId string, players []*Player) error {
	guild, err := session.Guild(guildId)
	if err != nil {
		return err
	}

	return fillPlayerEmojis(guild.Emojis, players, func(_ *Player) error {
		// no-op fill function because we assume that all the emojis are synchronized... asynchronously
		return nil
	})
}

func SyncMinecraftAvatarsToEmoji(session *discordgo.Session, guildId string, players []*Player) error {
	guild, err := session.Guild(guildId)
	if err != nil {
		return err
	}

	return fillPlayerEmojis(guild.Emojis, players, fillEmoji(session, guildId))
}

func checkIfEmojiNeedsUpdate(emojiId string, face []byte) bool {
	// if no id then that means emoji doesnt exist and we should create
	if emojiId == "" {
		return true
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png?quality=lossless", emojiId), nil)
	if err != nil {
		log.Printf("error preparing to fetch emoji from discord: %s", err.Error())
		return true
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error fetching emoji from discord: %s", err.Error())
		return true
	}
	defer resp.Body.Close()

	existingEmoji, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading emoji from discord: %s", err.Error())
		return true
	}

	return !bytes.Equal(existingEmoji, face)
}
