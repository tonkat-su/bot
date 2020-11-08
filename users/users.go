package users

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/tonkat-su/bot/mcuser"
)

type Service struct {
	Redis *redis.Client
}

func New(ctx context.Context, redisUrl string) (*Service, error) {
	redisOptions, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	redis := redis.NewClient(redisOptions)
	_, err = redis.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Service{Redis: redis}, nil
}

// if MinecraftUserId and MinecraftUsername are both not-nil, username is ignored and we store the uuid
// if MinecraftUserId is nil and MinecraftUsername is not-nil, we call a service to look up the uuid and store that
type RegisterInput struct {
	MinecraftUserId   *string
	MinecraftUsername *string
	DiscordUserId     string
}

func (input *RegisterInput) redisInput() map[string]interface{} {
	return map[string]interface{}{
		*input.MinecraftUserId: input.DiscordUserId,
		input.DiscordUserId:    *input.MinecraftUserId,
	}
}

func (svc *Service) Register(ctx context.Context, input *RegisterInput) error {
	if input == nil {
		return errors.New("registration input cannot be nil")
	}

	if input.MinecraftUserId == nil && input.MinecraftUsername == nil {
		return errors.New("one of MinecraftUserId or MinecraftUsername is required")
	}

	if input.DiscordUserId == "" {
		return errors.New("DiscordUserId is required")
	}

	if input.MinecraftUserId == nil && input.MinecraftUsername != nil {
		uuid, err := mcuser.GetUuid(*input.MinecraftUsername)
		if err != nil {
			return err
		}
		input.MinecraftUserId = &uuid
	}

	return svc.Redis.MSet(ctx, input.redisInput()).Err()
}
