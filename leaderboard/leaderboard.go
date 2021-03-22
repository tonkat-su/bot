package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type CloudwatchClient interface {
	cloudwatch.ListMetricsAPIClient
	cloudwatch.GetMetricDataAPIClient
	PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)
}

type Service struct {
	cloudwatch      CloudwatchClient
	namespacePrefix string
}

type Config struct {
	NamespacePrefix string
}

func New(awsCfg aws.Config, cfg *Config) (*Service, error) {
	return &Service{
		cloudwatch:      cloudwatch.NewFromConfig(awsCfg),
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

func (score *PlayerScore) metricDatum() types.MetricDatum {
	return types.MetricDatum{
		Dimensions: []types.Dimension{
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
		MetricData: make([]types.MetricDatum, len(input.Scores)),
	}

	for i, v := range input.Scores {
		if v.PlayerId == "" {
			return errors.New("leaderboard: got invalid player id")
		}
		metricInput.MetricData[i] = v.metricDatum()
	}

	_, err := svc.cloudwatch.PutMetricData(ctx, metricInput)
	if err != nil {
		return err
	}
	return nil
}

type Standings struct {
	SortedStandings []*PlayerScore
	LastUpdated     time.Time
}

func (svc *Service) fetchStandingsFromLastWeek(ctx context.Context, endTime time.Time) ([]types.MetricDataResult, error) {
	listMetricsInput := &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(svc.metricsNamespace()),
		MetricName: aws.String("PlayerScore"),
	}
	metrics := []types.Metric{}
	paginator := cloudwatch.NewListMetricsPaginator(svc.cloudwatch, listMetricsInput)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, output.Metrics...)
	}

	queries := make([]types.MetricDataQuery, len(metrics))
	for i, v := range metrics {
		var playerName string
		for _, dimension := range v.Dimensions {
			if *dimension.Name == "PlayerId" {
				playerName = *dimension.Value
			}
		}
		v := v
		queries[i] = types.MetricDataQuery{
			Id:    aws.String(fmt.Sprintf("query%d", i)),
			Label: &playerName,
			MetricStat: &types.MetricStat{
				Metric: &v,
				Period: aws.Int32(21600),
				Stat:   aws.String("Sum"),
			},
		}
	}

	getMetricDataInput := &cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(endTime),
		StartTime:         aws.Time(endTime.Add(-1 * 7 * 24 * time.Hour)),
		MetricDataQueries: queries,
	}
	results := []types.MetricDataResult{}
	getPaginator := cloudwatch.NewGetMetricDataPaginator(svc.cloudwatch, getMetricDataInput)
	for getPaginator.HasMorePages() {
		output, err := getPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, output.MetricDataResults...)
	}
	return results, nil
}

func transformCloudwatchResultsToStandings(results []types.MetricDataResult) *Standings {
	standings := &Standings{
		SortedStandings: make([]*PlayerScore, 0, len(results)),
	}

	for _, v := range results {
		score := &PlayerScore{
			PlayerId: *v.Label,
		}
		for _, value := range v.Values {
			score.Score += int64(value)
		}

		if score.Score > 0 {
			standings.SortedStandings = append(standings.SortedStandings, score)
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
