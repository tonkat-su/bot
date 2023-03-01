package main

import (
	"log"

	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name: "whitelist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:     discordgo.ApplicationCommandOptionSubCommand,
				Name:     "add",
				Required: true,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:     discordgo.ApplicationCommandOptionString,
						Name:     "username",
						Required: true,
					},
				},
			},
			{
				Type:     discordgo.ApplicationCommandOptionSubCommand,
				Name:     "remove",
				Required: true,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:     discordgo.ApplicationCommandOptionString,
						Name:     "username",
						Required: true,
					},
				},
			},
			{
				Type:     discordgo.ApplicationCommandOptionSubCommand,
				Name:     "list",
				Required: true,
			},
		},
	},
}

type Config struct {
	DiscordToken string `split_words:"true" required:"true"`
	GuildId      string `split_words:"true" required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("error reading envconfig: %s", err.Error())
	}

	client, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("error initializing discord client %s", err.Error())
	}

	defer func() {
		err = client.Close()
		if err != nil {
			log.Printf("error closing client %s", err.Error())
		}
	}()

	_, err = client.ApplicationCommandBulkOverwrite(client.State.User.ID, cfg.GuildId, commands)
	if err != nil {
		log.Fatalf("error registering commands %s", err.Error())
	}
}
