// Code generated by smithy-go-codegen DO NOT EDIT.

package secretsmanager

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Cancels the scheduled deletion of a secret by removing the DeletedDate time
// stamp. You can access a secret again after it has been restored. Secrets Manager
// generates a CloudTrail log entry when you call this action. Do not include
// sensitive information in request parameters because it might be logged. For more
// information, see Logging Secrets Manager events with CloudTrail
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieve-ct-entries.html).
// Required permissions: secretsmanager:RestoreSecret. For more information, see
// IAM policy actions for Secrets Manager
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_iam-permissions.html#reference_iam-permissions_actions)
// and Authentication and access control in Secrets Manager
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access.html).
func (c *Client) RestoreSecret(ctx context.Context, params *RestoreSecretInput, optFns ...func(*Options)) (*RestoreSecretOutput, error) {
	if params == nil {
		params = &RestoreSecretInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "RestoreSecret", params, optFns, c.addOperationRestoreSecretMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*RestoreSecretOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type RestoreSecretInput struct {

	// The ARN or name of the secret to restore. For an ARN, we recommend that you
	// specify a complete ARN rather than a partial ARN. See Finding a secret from a
	// partial ARN
	// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/troubleshoot.html#ARN_secretnamehyphen).
	//
	// This member is required.
	SecretId *string

	noSmithyDocumentSerde
}

type RestoreSecretOutput struct {

	// The ARN of the secret that was restored.
	ARN *string

	// The name of the secret that was restored.
	Name *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationRestoreSecretMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpRestoreSecret{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpRestoreSecret{}, middleware.After)
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
	if err = addOpRestoreSecretValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opRestoreSecret(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opRestoreSecret(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "RestoreSecret",
	}
}
