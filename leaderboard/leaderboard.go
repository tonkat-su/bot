package leaderboard

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/go-redis/redis/v8"
)

type Service struct {
	redis      *redis.Client
	cloudwatch cloudwatchiface.CloudWatchAPI
}

func New(ctx context.Context, redisUrl string, session *session.Session) (*Service, error) {
	redisOptions, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	redis := redis.NewClient(redisOptions)
	_, err = redis.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Service{
		redis:      redis,
		cloudwatch: cloudwatch.New(session),
	}, nil
}

type RecordScoresInput struct {
	Scores []*PlayerScore
}

type PlayerScore struct {
	PlayerId string
	Score    int64
}

func (score *PlayerScore) metricDatum() *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("PlayerId"),
				Value: aws.String(score.PlayerId),
			},
		},
		MetricName: aws.String("PlayerScore"),
		Value:      aws.Float64(float64(score.Score)),
	}
}

func (svc *Service) RecordScores(ctx context.Context, input *RecordScoresInput) error {
	metricInput := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("pumpcraft/Leaderboard"),
		MetricData: make([]*cloudwatch.MetricDatum, len(input.Scores)),
	}

	for i, v := range input.Scores {
		if v.PlayerId == "" {
			return errors.New("leaderboard: got invalid player id")
		}
		metricInput.MetricData[i] = v.metricDatum()
		if err := svc.redis.IncrBy(ctx, v.PlayerId, v.Score).Err(); err != nil {
			return fmt.Errorf("leaderboard: error incrementing player '%s' score: %s", v.PlayerId, err)
		}
	}

	_, err := svc.cloudwatch.PutMetricDataWithContext(ctx, metricInput)
	if err != nil {
		return err
	}
	return nil
}
