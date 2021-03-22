package users

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

type Service struct {
	DB        DynamoDBClient
	TableName string
}

func New(awsCfg aws.Config, tableName string) (*Service, error) {
	return &Service{
		DB:        dynamodb.NewFromConfig(awsCfg),
		TableName: tableName,
	}, nil
}

type RegisterInput struct {
	MinecraftUserId string
	DiscordUserId   string
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

	av, err := attributevalue.MarshalMap(input)
	if err != nil {
		return err
	}
	_, err = svc.DB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(svc.TableName),
		Item:      av,
	})
	return err
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
	result, err := svc.DB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(svc.TableName),
		Key: map[string]types.AttributeValue{
			"DiscordUserId": &types.AttributeValueMemberS{
				Value: input.Id,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var output LookupOutput
	err = attributevalue.UnmarshalMap(result.Item, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (svc *Service) LookupByMinecraftId(ctx context.Context, input *LookupInput) (*LookupOutput, error) {
	if input == nil {
		return nil, errors.New("input must not be nil")
	}

	expr, err := expression.NewBuilder().WithKeyCondition(expression.KeyEqual(expression.Key("MinecraftUserId"), expression.Value(input.Id))).Build()
	if err != nil {
		return nil, err
	}

	result, err := svc.DB.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(svc.TableName),
		IndexName:                 aws.String("MinecraftIdIndex"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) != 1 {
		return nil, nil
	}

	var output LookupOutput
	err = attributevalue.UnmarshalMap(result.Items[0], &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
