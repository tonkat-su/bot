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

// Creates a dashboard if it does not already exist, or updates an existing
// dashboard. If you update a dashboard, the entire contents are replaced with what
// you specify here. All dashboards in your account are global, not
// region-specific. A simple way to create a dashboard using PutDashboard is to
// copy an existing dashboard. To copy an existing dashboard using the console, you
// can load the dashboard and then use the View/edit source command in the Actions
// menu to display the JSON block for that dashboard. Another way to copy a
// dashboard is to use GetDashboard , and then use the data returned within
// DashboardBody as the template for the new dashboard when you call PutDashboard .
// When you create a dashboard with PutDashboard , a good practice is to add a text
// widget at the top of the dashboard with a message that the dashboard was created
// by script and should not be changed in the console. This message could also
// point console users to the location of the DashboardBody script or the
// CloudFormation template used to create the dashboard.
func (c *Client) PutDashboard(ctx context.Context, params *PutDashboardInput, optFns ...func(*Options)) (*PutDashboardOutput, error) {
	if params == nil {
		params = &PutDashboardInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "PutDashboard", params, optFns, c.addOperationPutDashboardMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*PutDashboardOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type PutDashboardInput struct {

	// The detailed information about the dashboard in JSON format, including the
	// widgets to include and their location on the dashboard. This parameter is
	// required. For more information about the syntax, see Dashboard Body Structure
	// and Syntax (https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/CloudWatch-Dashboard-Body-Structure.html)
	// .
	//
	// This member is required.
	DashboardBody *string

	// The name of the dashboard. If a dashboard with this name already exists, this
	// call modifies that dashboard, replacing its current contents. Otherwise, a new
	// dashboard is created. The maximum length is 255, and valid characters are A-Z,
	// a-z, 0-9, "-", and "_". This parameter is required.
	//
	// This member is required.
	DashboardName *string

	noSmithyDocumentSerde
}

type PutDashboardOutput struct {

	// If the input for PutDashboard was correct and the dashboard was successfully
	// created or modified, this result is empty. If this result includes only warning
	// messages, then the input was valid enough for the dashboard to be created or
	// modified, but some elements of the dashboard might not render. If this result
	// includes error messages, the input was not valid and the operation failed.
	DashboardValidationMessages []types.DashboardValidationMessage

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationPutDashboardMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsquery_serializeOpPutDashboard{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpPutDashboard{}, middleware.After)
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
	if err = addOpPutDashboardValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opPutDashboard(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opPutDashboard(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "monitoring",
		OperationName: "PutDashboard",
	}
}
