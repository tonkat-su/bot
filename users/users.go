package users

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Service struct {
	DB        dynamodbiface.DynamoDBAPI
	TableName string
}

func New(sess *session.Session, tableName string) (*Service, error) {
	return &Service{
		DB:        dynamodb.New(sess),
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

	av, err := dynamodbattribute.MarshalMap(input)
	if err != nil {
		return err
	}
	_, err = svc.DB.PutItemWithContext(ctx, &dynamodb.PutItemInput{
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
	result, err := svc.DB.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(svc.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"DiscordUserId": {
				S: aws.String(input.Id),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var output LookupOutput
	err = dynamodbattribute.UnmarshalMap(result.Item, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (svc *Service) LookupByMinecraftId(ctx context.Context, input *LookupInput) (*LookupOutput, error) {
	if input == nil {
		return nil, errors.New("input must not be nil")
	}
	result, err := svc.DB.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(svc.TableName),
		IndexName:              aws.String("MinecraftIdIndex"),
		KeyConditionExpression: aws.String("MinecraftUserId = :user_id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":user_id": {S: aws.String(input.Id)},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) != 1 {
		return nil, nil
	}

	var output LookupOutput
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
