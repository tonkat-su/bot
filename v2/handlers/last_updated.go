package handlers

import (
	"time"

	"github.com/andersfylling/disgord"
)

func AppendLastUpdatedEmbedField(fields []*disgord.EmbedField, timestamp time.Time) ([]*disgord.EmbedField, error) {
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, err
	}
	return append(fields, &disgord.EmbedField{
		Name:  "*last updated:*",
		Value: timestamp.In(tz).Format(time.UnixDate),
	}), nil
}
