// Code generated by smithy-go-codegen DO NOT EDIT.

package secretsmanager

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Retrieves the details of a secret. It does not include the encrypted fields.
// Secrets Manager only returns fields populated with a value in the response.
// Minimum permissions To run this command, you must have the following
// permissions:
//
// * secretsmanager:DescribeSecret
//
// Related operations
//
// * To create a
// secret, use CreateSecret.
//
// * To modify a secret, use UpdateSecret.
//
// * To
// retrieve the encrypted secret information in a version of the secret, use
// GetSecretValue.
//
// * To list all of the secrets in the AWS account, use
// ListSecrets.
func (c *Client) DescribeSecret(ctx context.Context, params *DescribeSecretInput, optFns ...func(*Options)) (*DescribeSecretOutput, error) {
	if params == nil {
		params = &DescribeSecretInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribeSecret", params, optFns, addOperationDescribeSecretMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribeSecretOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribeSecretInput struct {

	// The identifier of the secret whose details you want to retrieve. You can specify
	// either the Amazon Resource Name (ARN) or the friendly name of the secret. If you
	// specify an ARN, we generally recommend that you specify a complete ARN. You can
	// specify a partial ARN too—for example, if you don’t include the final hyphen and
	// six random characters that Secrets Manager adds at the end of the ARN when you
	// created the secret. A partial ARN match can work as long as it uniquely matches
	// only one secret. However, if your secret has a name that ends in a hyphen
	// followed by six characters (before Secrets Manager adds the hyphen and six
	// characters to the ARN) and you try to use that as a partial ARN, then those
	// characters cause Secrets Manager to assume that you’re specifying a complete
	// ARN. This confusion can cause unexpected results. To avoid this situation, we
	// recommend that you don’t create secret names ending with a hyphen followed by
	// six characters. If you specify an incomplete ARN without the random suffix, and
	// instead provide the 'friendly name', you must not include the random suffix. If
	// you do include the random suffix added by Secrets Manager, you receive either a
	// ResourceNotFoundException or an AccessDeniedException error, depending on your
	// permissions.
	//
	// This member is required.
	SecretId *string
}

type DescribeSecretOutput struct {

	// The ARN of the secret.
	ARN *string

	// The date you created the secret.
	CreatedDate *time.Time

	// This value exists if the secret is scheduled for deletion. Some time after the
	// specified date and time, Secrets Manager deletes the secret and all of its
	// versions. If a secret is scheduled for deletion, then its details, including the
	// encrypted secret information, is not accessible. To cancel a scheduled deletion
	// and restore access, use RestoreSecret.
	DeletedDate *time.Time

	// The user-provided description of the secret.
	Description *string

	// The ARN or alias of the AWS KMS customer master key (CMK) that's used to encrypt
	// the SecretString or SecretBinary fields in each version of the secret. If you
	// don't provide a key, then Secrets Manager defaults to encrypting the secret
	// fields with the default AWS KMS CMK (the one named awssecretsmanager) for this
	// account.
	KmsKeyId *string

	// The last date that this secret was accessed. This value is truncated to midnight
	// of the date and therefore shows only the date, not the time.
	LastAccessedDate *time.Time

	// The last date and time that this secret was modified in any way.
	LastChangedDate *time.Time

	// The last date and time that the rotation process for this secret was invoked.
	// The most recent date and time that the Secrets Manager rotation process
	// successfully completed. If the secret doesn't rotate, Secrets Manager returns a
	// null value.
	LastRotatedDate *time.Time

	// The user-provided friendly name of the secret.
	Name *string

	// Returns the name of the service that created this secret.
	OwningService *string

	// Specifies the primary region for secret replication.
	PrimaryRegion *string

	// Describes a list of replication status objects as InProgress, Failed or InSync.P
	ReplicationStatus []types.ReplicationStatusType

	// Specifies whether automatic rotation is enabled for this secret. To enable
	// rotation, use RotateSecret with AutomaticallyRotateAfterDays set to a value
	// greater than 0. To disable rotation, use CancelRotateSecret.
	RotationEnabled bool

	// The ARN of a Lambda function that's invoked by Secrets Manager to rotate the
	// secret either automatically per the schedule or manually by a call to
	// RotateSecret.
	RotationLambdaARN *string

	// A structure with the rotation configuration for this secret.
	RotationRules *types.RotationRulesType

	// The list of user-defined tags that are associated with the secret. To add tags
	// to a secret, use TagResource. To remove tags, use UntagResource.
	Tags []types.Tag

	// A list of all of the currently assigned VersionStage staging labels and the
	// VersionId that each is attached to. Staging labels are used to keep track of the
	// different versions during the rotation process. A version that does not have any
	// staging labels attached is considered deprecated and subject to deletion. Such
	// versions are not included in this list.
	VersionIdsToStages map[string][]string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addOperationDescribeSecretMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpDescribeSecret{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpDescribeSecret{}, middleware.After)
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
	if err = addOpDescribeSecretValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribeSecret(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opDescribeSecret(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "DescribeSecret",
	}
}
