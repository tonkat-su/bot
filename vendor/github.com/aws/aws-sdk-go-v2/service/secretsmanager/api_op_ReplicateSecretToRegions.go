// Code generated by smithy-go-codegen DO NOT EDIT.

package secretsmanager

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Replicates the secret to a new Regions. See Multi-Region secrets
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/create-manage-multi-region-secrets.html).
// Required permissions: secretsmanager:ReplicateSecretToRegions. For more
// information, see  IAM policy actions for Secrets Manager
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_iam-permissions.html#reference_iam-permissions_actions)
// and Authentication and access control in Secrets Manager
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access.html).
func (c *Client) ReplicateSecretToRegions(ctx context.Context, params *ReplicateSecretToRegionsInput, optFns ...func(*Options)) (*ReplicateSecretToRegionsOutput, error) {
	if params == nil {
		params = &ReplicateSecretToRegionsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ReplicateSecretToRegions", params, optFns, c.addOperationReplicateSecretToRegionsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ReplicateSecretToRegionsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ReplicateSecretToRegionsInput struct {

	// A list of Regions in which to replicate the secret.
	//
	// This member is required.
	AddReplicaRegions []types.ReplicaRegionType

	// The ARN or name of the secret to replicate.
	//
	// This member is required.
	SecretId *string

	// Specifies whether to overwrite a secret with the same name in the destination
	// Region.
	ForceOverwriteReplicaSecret bool

	noSmithyDocumentSerde
}

type ReplicateSecretToRegionsOutput struct {

	// The ARN of the primary secret.
	ARN *string

	// The status of replication.
	ReplicationStatus []types.ReplicationStatusType

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationReplicateSecretToRegionsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpReplicateSecretToRegions{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpReplicateSecretToRegions{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpReplicateSecretToRegionsValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opReplicateSecretToRegions(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opReplicateSecretToRegions(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "ReplicateSecretToRegions",
	}
}
