package users

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
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
	MinecraftUserId string
	DiscordUserId   string
}

func (input *RegisterInput) redisInput() map[string]interface{} {
	return map[string]interface{}{
		minecraftUserIdRedisKey(input.MinecraftUserId): input.DiscordUserId,
		discordUserIdRedisKey(input.DiscordUserId):     input.MinecraftUserId,
	}
}

func (svc *Service) Register(ctx context.Context, input *RegisterInput) error {
	if input == nil {
		return errors.New("registration input cannot be nil")
	}

	if input.MinecraftUserId == "" {
		return errors.New("MinecraftUserId is required")
	}

	if input.DiscordUserId == "" {
		return errors.New("DiscordUserId is required")
	}

	return svc.Redis.MSet(ctx, input.redisInput()).Err()
}

type LookupInput struct {
	Id string
}

type LookupOutput struct {
	MinecraftUserId string
	DiscordUserId   string
}

func (svc *Service) LookupByDiscordId(ctx context.Context, input *LookupInput) (*LookupOutput, error) {
	if input == nil {
		return nil, errors.New("input must not be nil")
	}
	result, err := svc.Redis.Get(ctx, discordUserIdRedisKey(input.Id)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return &LookupOutput{
		DiscordUserId:   input.Id,
		MinecraftUserId: result,
	}, nil
}

func (svc *Service) LookupByMinecraftId(ctx context.Context, input *LookupInput) (*LookupOutput, error) {
	if input == nil {
		return nil, errors.New("input must not be nil")
	}

	result, err := svc.Redis.Get(ctx, minecraftUserIdRedisKey(input.Id)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return &LookupOutput{
		DiscordUserId:   result,
		MinecraftUserId: input.Id,
	}, nil
}

func minecraftUserIdRedisKey(minecraftId string) string {
	return "minecraft-user-id:" + minecraftId
}

func discordUserIdRedisKey(discordId string) string {
	return "discord-user-id:" + discordId
}
