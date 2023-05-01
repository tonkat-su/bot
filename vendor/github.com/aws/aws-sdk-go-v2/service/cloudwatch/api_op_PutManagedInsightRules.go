// Code generated by smithy-go-codegen DO NOT EDIT.

package cloudwatch

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Creates a managed Contributor Insights rule for a specified Amazon Web Services
// resource. When you enable a managed rule, you create a Contributor Insights rule
// that collects data from Amazon Web Services services. You cannot edit these
// rules with PutInsightRule . The rules can be enabled, disabled, and deleted
// using EnableInsightRules , DisableInsightRules , and DeleteInsightRules . If a
// previously created managed rule is currently disabled, a subsequent call to this
// API will re-enable it. Use ListManagedInsightRules to describe all available
// rules.
func (c *Client) PutManagedInsightRules(ctx context.Context, params *PutManagedInsightRulesInput, optFns ...func(*Options)) (*PutManagedInsightRulesOutput, error) {
	if params == nil {
		params = &PutManagedInsightRulesInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "PutManagedInsightRules", params, optFns, c.addOperationPutManagedInsightRulesMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*PutManagedInsightRulesOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type PutManagedInsightRulesInput struct {

	// A list of ManagedRules to enable.
	//
	// This member is required.
	ManagedRules []types.ManagedRule

	noSmithyDocumentSerde
}

type PutManagedInsightRulesOutput struct {

	// An array that lists the rules that could not be enabled.
	Failures []types.PartialFailure

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationPutManagedInsightRulesMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsquery_serializeOpPutManagedInsightRules{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpPutManagedInsightRules{}, middleware.After)
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
	if err = addOpPutManagedInsightRulesValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opPutManagedInsightRules(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
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

func newServiceMetadataMiddleware_opPutManagedInsightRules(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "monitoring",
		OperationName: "PutManagedInsightRules",
	}
}
