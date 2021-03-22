// Code generated by smithy-go-codegen DO NOT EDIT.

package secretsmanager

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Lists all of the secrets that are stored by Secrets Manager in the AWS account.
// To list the versions currently stored for a specific secret, use
// ListSecretVersionIds. The encrypted fields SecretString and SecretBinary are not
// included in the output. To get that information, call the GetSecretValue
// operation. Always check the NextToken response parameter when calling any of the
// List* operations. These operations can occasionally return an empty or shorter
// than expected list of results even when there more results become available.
// When this happens, the NextToken response parameter contains a value to pass to
// the next call to the same API to request the next part of the list. Minimum
// permissions To run this command, you must have the following permissions:
//
// *
// secretsmanager:ListSecrets
//
// Related operations
//
// * To list the versions attached
// to a secret, use ListSecretVersionIds.
func (c *Client) ListSecrets(ctx context.Context, params *ListSecretsInput, optFns ...func(*Options)) (*ListSecretsOutput, error) {
	if params == nil {
		params = &ListSecretsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListSecrets", params, optFns, addOperationListSecretsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListSecretsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListSecretsInput struct {

	// Lists the secret request filters.
	Filters []types.Filter

	// (Optional) Limits the number of results you want to include in the response. If
	// you don't include this parameter, it defaults to a value that's specific to the
	// operation. If additional items exist beyond the maximum you specify, the
	// NextToken response element is present and has a value (isn't null). Include that
	// value as the NextToken request parameter in the next call to the operation to
	// get the next part of the results. Note that Secrets Manager might return fewer
	// results than the maximum even when there are more results available. You should
	// check NextToken after every operation to ensure that you receive all of the
	// results.
	MaxResults int32

	// (Optional) Use this parameter in a request if you receive a NextToken response
	// in a previous request indicating there's more output available. In a subsequent
	// call, set it to the value of the previous call NextToken response to indicate
	// where the output should continue from.
	NextToken *string

	// Lists secrets in the requested order.
	SortOrder types.SortOrderType
}

type ListSecretsOutput struct {

	// If present in the response, this value indicates that there's more output
	// available than included in the current response. This can occur even when the
	// response includes no values at all, such as when you ask for a filtered view of
	// a very long list. Use this value in the NextToken request parameter in a
	// subsequent call to the operation to continue processing and get the next part of
	// the output. You should repeat this until the NextToken response element comes
	// back empty (as null).
	NextToken *string

	// A list of the secrets in the account.
	SecretList []types.SecretListEntry

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addOperationListSecretsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpListSecrets{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpListSecrets{}, middleware.After)
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListSecrets(options.Region), middleware.Before); err != nil {
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

// ListSecretsAPIClient is a client that implements the ListSecrets operation.
type ListSecretsAPIClient interface {
	ListSecrets(context.Context, *ListSecretsInput, ...func(*Options)) (*ListSecretsOutput, error)
}

var _ ListSecretsAPIClient = (*Client)(nil)

// ListSecretsPaginatorOptions is the paginator options for ListSecrets
type ListSecretsPaginatorOptions struct {
	// (Optional) Limits the number of results you want to include in the response. If
	// you don't include this parameter, it defaults to a value that's specific to the
	// operation. If additional items exist beyond the maximum you specify, the
	// NextToken response element is present and has a value (isn't null). Include that
	// value as the NextToken request parameter in the next call to the operation to
	// get the next part of the results. Note that Secrets Manager might return fewer
	// results than the maximum even when there are more results available. You should
	// check NextToken after every operation to ensure that you receive all of the
	// results.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListSecretsPaginator is a paginator for ListSecrets
type ListSecretsPaginator struct {
	options   ListSecretsPaginatorOptions
	client    ListSecretsAPIClient
	params    *ListSecretsInput
	nextToken *string
	firstPage bool
}

// NewListSecretsPaginator returns a new ListSecretsPaginator
func NewListSecretsPaginator(client ListSecretsAPIClient, params *ListSecretsInput, optFns ...func(*ListSecretsPaginatorOptions)) *ListSecretsPaginator {
	if params == nil {
		params = &ListSecretsInput{}
	}

	options := ListSecretsPaginatorOptions{}
	if params.MaxResults != 0 {
		options.Limit = params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListSecretsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListSecretsPaginator) HasMorePages() bool {
	return p.firstPage || p.nextToken != nil
}

// NextPage retrieves the next ListSecrets page.
func (p *ListSecretsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListSecretsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	params.MaxResults = p.options.Limit

	result, err := p.client.ListSecrets(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken && prevToken != nil && p.nextToken != nil && *prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

func newServiceMetadataMiddleware_opListSecrets(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "ListSecrets",
	}
}
