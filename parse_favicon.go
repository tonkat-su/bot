package main

import (
	"bytes"

	"github.com/bwmarrin/discordgo"
	"github.com/vincent-petithory/dataurl"
)

func parseFavIcon(server string, u string) (file *discordgo.File, err error) {
	dataURL, err := dataurl.DecodeString(u)
	if err != nil {
		return nil, err
	}
	return &discordgo.File{
		Name:        server,
		ContentType: dataURL.MediaType.ContentType(),
		Reader:      bytes.NewReader(dataURL.Data),
	}, nil
}
