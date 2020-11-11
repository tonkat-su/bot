package handlers

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func AppendLastUpdatedEmbedField(fields []*discordgo.MessageEmbedField, timestamp time.Time) ([]*discordgo.MessageEmbedField, error) {
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, err
	}
	return append(fields, &discordgo.MessageEmbedField{
		Name:  "*last updated:*",
		Value: timestamp.In(tz).Format(time.UnixDate),
	}), nil
}
