package interactions

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tonkat-su/bot/imgur"
)

type Config struct {
	ImgurClientId string `split_words:"true" required:"true"`

	DiscordToken         string `split_words:"true" required:"true"`
	DiscordWebhookPubkey string `split_words:"true" required:"true"`

	MinecraftServerName string `split_words:"true" required:"true"`
	MinecraftServerHost string `split_words:"true" required:"true"`
	RconPassword        string `split_words:"true" required:"true"`
	RconHostport        string `split_words:"true" required:"true"`

	DiscordGuildId string `split_words:"true" required:"true"`
}

func NewServer(cfg *Config) (*Server, error) {
	discordClient, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, err
	}

	discordClient.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildEmojis

	srv := &Server{
		s:   discordClient,
		cfg: cfg,
		imgur: &imgur.Client{
			ClientId: cfg.ImgurClientId,
		},
	}

	srv.handlers = map[string]InteractionHandler{
		"online":    srv.online,
		"whitelist": srv.whitelist,
		"test":      srv.test,
		"version":   srv.version,
	}

	discordClient.AddHandler(srv.onReady)

	err = discordClient.Open()
	if err != nil {
		return nil, err
	}

	return srv, nil
}

type Server struct {
	s        *discordgo.Session
	cfg      *Config
	imgur    *imgur.Client
	handlers map[string]InteractionHandler
}

func (srv *Server) Close() error {
	return srv.s.Close()
}

type InteractionHandler func(http.ResponseWriter, discordgo.Interaction, *discordgo.Session)

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	discordPubkey, err := decodeDiscordWebhookPubkey(srv.cfg.DiscordWebhookPubkey)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// validate signature from discord and return 401 if invalid
	if !discordgo.VerifyInteraction(r, discordPubkey) {
		writeResponse(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	// marshal interaction webhook data
	var event discordgo.Interaction
	err = json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.Printf("invalid interaction data: %s", err.Error())
		writeResponse(w, http.StatusBadRequest, "error decoding json request")
		return
	}

	switch event.Type {
	case discordgo.InteractionPing:
		// reply with a pong when discord pings us
		log.Println("ping received")
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"type":1}`)
		if err != nil {
			log.Printf("error ACKing ping: %s", err.Error())
		}
		return
	case discordgo.InteractionApplicationCommand:
		srv.routeApplicationCommand(w, event)
		return
	case discordgo.InteractionMessageComponent:
		writeResponse(w, http.StatusUnprocessableEntity, "message interaction not implemented yet")
		return
	}

	writeResponse(w, http.StatusUnprocessableEntity, "invalid event type")
}

func (srv *Server) routeApplicationCommand(w http.ResponseWriter, event discordgo.Interaction) {
	data := event.ApplicationCommandData()
	if handler, ok := srv.handlers[data.Name]; ok {
		handler(w, event, srv.s)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: body,
		},
	})
	if err != nil {
		log.Printf("failed to encode body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func replyToInteraction(request discordgo.Interaction, response discordgo.InteractionResponse) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(response)
	if err != nil {
		return err
	}

	_, err = http.Post(fmt.Sprintf("https://discord.com/api/v8/interactions/%s/%s/callback", request.ID, request.Token), "application/json", &buf)
	if err != nil {
		return err
	}
	return nil
}

func decodeDiscordWebhookPubkey(k string) (ed25519.PublicKey, error) {
	data, err := hex.DecodeString(k)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(data), nil
}

func (srv *Server) onReady(s *discordgo.Session, event *discordgo.Ready) {
	guilds := []string{}
	for _, guild := range event.Guilds {
		guilds = append(guilds, guild.Name)
	}
	log.Printf("guilds joined: %s", strings.Join(guilds, ", "))
}
