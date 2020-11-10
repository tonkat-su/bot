package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

type Service struct {
	cloudwatch      cloudwatchiface.CloudWatchAPI
	namespacePrefix string
}

type Config struct {
	NamespacePrefix string
}

func New(session *session.Session, cfg *Config) (*Service, error) {
	return &Service{
		cloudwatch:      cloudwatch.New(session),
		namespacePrefix: cfg.NamespacePrefix,
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

func (svc *Service) metricsNamespace() string {
	return filepath.Join(svc.namespacePrefix, "Leaderboard")
}

func (svc *Service) RecordScores(ctx context.Context, input *RecordScoresInput) error {
	metricInput := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(svc.metricsNamespace()),
		MetricData: make([]*cloudwatch.MetricDatum, len(input.Scores)),
	}

	for i, v := range input.Scores {
		if v.PlayerId == "" {
			return errors.New("leaderboard: got invalid player id")
		}
		metricInput.MetricData[i] = v.metricDatum()
	}

	_, err := svc.cloudwatch.PutMetricDataWithContext(ctx, metricInput)
	if err != nil {
		return err
	}
	return nil
}

type Standings struct {
	SortedStandings []*PlayerScore
	LastUpdated     time.Time
}

func (svc *Service) fetchStandingsFromLastWeek(ctx context.Context, endTime time.Time) ([]*cloudwatch.MetricDataResult, error) {
	listMetricsInput := &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(svc.metricsNamespace()),
		MetricName: aws.String("PlayerScore"),
	}
	metrics := []*cloudwatch.Metric{}
	err := svc.cloudwatch.ListMetricsPagesWithContext(ctx, listMetricsInput, func(output *cloudwatch.ListMetricsOutput, more bool) bool {
		metrics = append(metrics, output.Metrics...)
		return more
	})
	if err != nil {
		return nil, err
	}

	queries := make([]*cloudwatch.MetricDataQuery, len(metrics))
	for i, v := range metrics {
		var playerName *string
		for _, dimension := range v.Dimensions {
			if aws.StringValue(dimension.Name) == "PlayerId" {
				playerName = dimension.Value
			}
		}
		queries[i] = &cloudwatch.MetricDataQuery{
			Id:    aws.String(fmt.Sprintf("query%d", i)),
			Label: playerName,
			MetricStat: &cloudwatch.MetricStat{
				Metric: v,
				Period: aws.Int64(21600),
				Stat:   aws.String("Sum"),
			},
		}
	}

	getMetricDataInput := &cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(endTime),
		StartTime:         aws.Time(endTime.Add(-1 * 7 * 24 * time.Hour)),
		MetricDataQueries: queries,
	}
	results := []*cloudwatch.MetricDataResult{}
	err = svc.cloudwatch.GetMetricDataPagesWithContext(ctx, getMetricDataInput, func(output *cloudwatch.GetMetricDataOutput, more bool) bool {
		results = append(results, output.MetricDataResults...)
		return more
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func transformCloudwatchResultsToStandings(results []*cloudwatch.MetricDataResult) *Standings {
	standings := &Standings{
		SortedStandings: make([]*PlayerScore, len(results)),
	}
	for i, v := range results {
		standings.SortedStandings[i] = &PlayerScore{
			PlayerId: aws.StringValue(v.Label),
		}
		for _, value := range v.Values {
			standings.SortedStandings[i].Score += int64(aws.Float64Value(value))
		}
	}
	sort.SliceStable(standings.SortedStandings, func(i, j int) bool {
		return standings.SortedStandings[i].Score > standings.SortedStandings[j].Score
	})
	return standings
}

func (svc *Service) GetStandings(ctx context.Context) (*Standings, error) {
	endTime := time.Now().Round(5 * time.Minute)
	results, err := svc.fetchStandingsFromLastWeek(ctx, endTime)
	if err != nil {
		return nil, err
	}
	standings := transformCloudwatchResultsToStandings(results)
	standings.LastUpdated = endTime
	return standings, nil
}
