package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bsdlp/envconfig"
	"github.com/bwmarrin/discordgo"
	"github.com/jltobler/go-rcon"
	"github.com/tonkat-su/bot/emoji"
)

type Config struct {
	DiscordToken   string `split_words:"true" required:"true"`
	DiscordGuildId string `split_words:"true" required:"true"`

	MinecraftRconHostPort string `split_words:"true" required:"true"`
	MinecraftRconPassword string `split_words:"true" required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("error processing config: %s", err.Error())
	}

	discordClient, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("error creating discord client: %s", err.Error())
	}

	rconClient := rcon.NewClient("rcon://"+cfg.MinecraftRconHostPort, cfg.MinecraftRconPassword)

	lambda.Start(func(ctx context.Context) error {
		// get players in whitelist
		output, err := rconClient.Send("whitelist list")
		if err != nil {
			return err
		}

		// parse rcon output, looks something like:
		// There are 14 whitelisted players: ouroboronn, MuchJokes, ImBith, Tigglywuff, piecatjustice, Rainefan, Nomibby, Sharisi, scholtez, Afadra, seputus, odiistorm, piecat314, bsdlp
		players := strings.Split(strings.Split(output, ": ")[1], ", ")
		emojis := make([]*emoji.Player, len(players))
		for i, v := range players {
			emojis[i] = &emoji.Player{
				Name: v,
			}
		}

		return emoji.SyncMinecraftAvatarsToEmoji(discordClient, cfg.DiscordGuildId, emojis)
	})
}
