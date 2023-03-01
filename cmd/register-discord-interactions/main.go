package main

import (
	"log"

	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "whitelist",
		Description: "whitelist command",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "add minecraft user to whitelist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "username",
						Description: "minecraft username to add to whitelist",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "command to remove minecraft user from whitelist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "username",
						Description: "minecraft username to remove from whitelist",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "command to list users currently whitelisted",
			},
		},
	},
}

type Config struct {
	DiscordToken string `split_words:"true" required:"true"`
	GuildId      string `split_words:"true" required:"true"`
	AppId        string `split_words:"true" required:"true"`
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

	_, err = client.ApplicationCommandBulkOverwrite(cfg.AppId, cfg.GuildId, commands)
	if err != nil {
		log.Fatalf("error registering commands %s", err.Error())
	}
}
