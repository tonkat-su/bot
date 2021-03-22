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

// Attaches one or more tags, each consisting of a key name and a value, to the
// specified secret. Tags are part of the secret's overall metadata, and are not
// associated with any specific version of the secret. This operation only appends
// tags to the existing list of tags. To remove tags, you must use UntagResource.
// The following basic restrictions apply to tags:
//
// * Maximum number of tags per
// secret—50
//
// * Maximum key length—127 Unicode characters in UTF-8
//
// * Maximum value
// length—255 Unicode characters in UTF-8
//
// * Tag keys and values are case
// sensitive.
//
// * Do not use the aws: prefix in your tag names or values because AWS
// reserves it for AWS use. You can't edit or delete tag names or values with this
// prefix. Tags with this prefix do not count against your tags per secret
// limit.
//
// * If you use your tagging schema across multiple services and resources,
// remember other services might have restrictions on allowed characters. Generally
// allowed characters: letters, spaces, and numbers representable in UTF-8, plus
// the following special characters: + - = . _ : / @.
//
// If you use tags as part of
// your security strategy, then adding or removing a tag can change permissions. If
// successfully completing this operation would result in you losing your
// permissions for this secret, then the operation is blocked and returns an Access
// Denied error. Minimum permissions To run this command, you must have the
// following permissions:
//
// * secretsmanager:TagResource
//
// Related operations
//
// * To
// remove one or more tags from the collection attached to a secret, use
// UntagResource.
//
// * To view the list of tags attached to a secret, use
// DescribeSecret.
func (c *Client) TagResource(ctx context.Context, params *TagResourceInput, optFns ...func(*Options)) (*TagResourceOutput, error) {
	if params == nil {
		params = &TagResourceInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "TagResource", params, optFns, addOperationTagResourceMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*TagResourceOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type TagResourceInput struct {

	// The identifier for the secret that you want to attach tags to. You can specify
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

	// The tags to attach to the secret. Each element in the list consists of a Key and
	// a Value. This parameter to the API requires a JSON text string argument. For
	// information on how to format a JSON parameter for the various command line tool
	// environments, see Using JSON for Parameters
	// (https://docs.aws.amazon.com/cli/latest/userguide/cli-using-param.html#cli-using-param-json)
	// in the AWS CLI User Guide. For the AWS CLI, you can also use the syntax: --Tags
	// Key="Key1",Value="Value1" Key="Key2",Value="Value2"[,…]
	//
	// This member is required.
	Tags []types.Tag
}

type TagResourceOutput struct {
	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addOperationTagResourceMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpTagResource{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpTagResource{}, middleware.After)
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
	if err = addOpTagResourceValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opTagResource(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opTagResource(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "TagResource",
	}
}
