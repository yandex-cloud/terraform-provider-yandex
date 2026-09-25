package operation

import (
	"context"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Operation_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/conn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation/metadata"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

type (
	// Client is an operation service client for manage long operations in YDB
	//
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	Client struct {
		operationServiceClient Ydb_Operation_V1.OperationServiceClient
	}
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	listOperationsWithNextToken[PT metadata.Constraint[T], T metadata.TypesConstraint] struct {
		listOperations[PT, T]
		NextToken string
	}
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	listOperations[PT metadata.Constraint[T], T metadata.TypesConstraint] struct {
		Operations []*typedOperation[PT, T]
	}
	operation struct {
		ID            string
		Ready         bool
		Status        string
		ConsumedUnits float64
	}
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	typedOperation[PT metadata.Constraint[T], T metadata.TypesConstraint] struct {
		operation
		Metadata *T
	}
)

// Get returns operation status by ID
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) Get(ctx context.Context, opID string) (*operation, error) {
	op, err := get(ctx, c.operationServiceClient, opID)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return op, nil
}

func get(
	ctx context.Context, client Ydb_Operation_V1.OperationServiceClient, opID string,
) (*operation, error) {
	status, err := retry.RetryWithResult(ctx, func(ctx context.Context) (*operation, error) {
		response, err := client.GetOperation(
			conn.WithoutWrapping(ctx),
			&Ydb_Operations.GetOperationRequest{
				Id: opID,
			},
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		var md Ydb_Query.ExecuteScriptMetadata
		err = response.GetOperation().GetMetadata().UnmarshalTo(&md)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return &operation{
			Ready:  response.GetOperation().GetReady(),
			Status: response.GetOperation().GetStatus().String(),
		}, nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return status, nil
}

func list[PT metadata.Constraint[T], T metadata.TypesConstraint](
	ctx context.Context, client Ydb_Operation_V1.OperationServiceClient, request *Ydb_Operations.ListOperationsRequest,
) (*listOperationsWithNextToken[PT, T], error) {
	operations, err := retry.RetryWithResult(ctx, func(ctx context.Context) (
		operations *listOperationsWithNextToken[PT, T], _ error,
	) {
		response, err := client.ListOperations(conn.WithoutWrapping(ctx), request)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		if response.GetStatus() != Ydb.StatusIds_SUCCESS {
			return nil, xerrors.WithStackTrace(xerrors.Operation(
				xerrors.WithStatusCode(response.GetStatus()),
				xerrors.WithIssues(response.GetIssues()),
			))
		}

		operations = &listOperationsWithNextToken[PT, T]{
			listOperations: listOperations[PT, T]{
				Operations: make([]*typedOperation[PT, T], 0, len(response.GetOperations())),
			},
			NextToken: response.GetNextPageToken(),
		}

		for _, op := range response.GetOperations() {
			operations.Operations = append(operations.Operations, &typedOperation[PT, T]{
				operation: operation{
					ID:            op.GetId(),
					Ready:         op.GetReady(),
					Status:        op.GetStatus().String(),
					ConsumedUnits: op.GetCostInfo().GetConsumedUnits(),
				},
				Metadata: metadata.FromProto[PT, T](op.GetMetadata()),
			})
		}

		return operations, nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return operations, nil
}

func cancel(
	ctx context.Context, client Ydb_Operation_V1.OperationServiceClient, opID string,
) error {
	err := retry.Retry(ctx, func(ctx context.Context) error {
		response, err := client.CancelOperation(conn.WithoutWrapping(ctx), &Ydb_Operations.CancelOperationRequest{
			Id: opID,
		})
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		if response.GetStatus() != Ydb.StatusIds_SUCCESS {
			return xerrors.WithStackTrace(xerrors.Operation(
				xerrors.WithStatusCode(response.GetStatus()),
				xerrors.WithIssues(response.GetIssues()),
			))
		}

		return nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// Cancel starts cancellation of a long-running operation.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) Cancel(ctx context.Context, opID string) error {
	err := cancel(ctx, c.operationServiceClient, opID)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func forget(
	ctx context.Context, client Ydb_Operation_V1.OperationServiceClient, opID string,
) error {
	err := retry.Retry(ctx, func(ctx context.Context) error {
		response, err := client.ForgetOperation(conn.WithoutWrapping(ctx), &Ydb_Operations.ForgetOperationRequest{
			Id: opID,
		})
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		if response.GetStatus() != Ydb.StatusIds_SUCCESS {
			return xerrors.WithStackTrace(xerrors.Operation(
				xerrors.WithStatusCode(response.GetStatus()),
				xerrors.WithIssues(response.GetIssues()),
			))
		}

		return nil
	}, retry.WithIdempotent(true))
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// Forget forgets long-running operation. It does not cancel the operation and
// returns an error if operation was not completed.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) Forget(ctx context.Context, opID string) error {
	err := forget(ctx, c.operationServiceClient, opID)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) Close(ctx context.Context) error {
	return nil
}

// ListBuildIndex returns list of build index operations
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) ListBuildIndex(ctx context.Context) (
	*listOperations[*metadata.BuildIndex, metadata.BuildIndex], error,
) {
	request := &options.ListOperationsRequest{
		ListOperationsRequest: Ydb_Operations.ListOperationsRequest{
			Kind: kindBuildIndex,
		},
	}

	operations, err := list[*metadata.BuildIndex, metadata.BuildIndex](
		ctx, c.operationServiceClient, &request.ListOperationsRequest,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &operations.listOperations, nil
}

// ListImportFromS3 returns list of import from s3 operations
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) ListImportFromS3(ctx context.Context) (
	*listOperations[*metadata.ImportFromS3, metadata.ImportFromS3], error,
) {
	request := &options.ListOperationsRequest{
		ListOperationsRequest: Ydb_Operations.ListOperationsRequest{
			Kind: kindImportFromS3,
		},
	}

	operations, err := list[*metadata.ImportFromS3, metadata.ImportFromS3](
		ctx, c.operationServiceClient, &request.ListOperationsRequest,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &operations.listOperations, nil
}

// ListExportToS3 returns list of export to s3 operations
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) ListExportToS3(ctx context.Context) (
	*listOperations[*metadata.ExportToS3, metadata.ExportToS3], error,
) {
	request := &options.ListOperationsRequest{
		ListOperationsRequest: Ydb_Operations.ListOperationsRequest{
			Kind: kindExportToS3,
		},
	}

	operations, err := list[*metadata.ExportToS3, metadata.ExportToS3](
		ctx, c.operationServiceClient, &request.ListOperationsRequest,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &operations.listOperations, nil
}

// ListExportToYT returns list of export to YT operations
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) ListExportToYT(ctx context.Context) (
	*listOperations[*metadata.ExportToYT, metadata.ExportToYT], error,
) {
	request := &options.ListOperationsRequest{
		ListOperationsRequest: Ydb_Operations.ListOperationsRequest{
			Kind: kindExportToYT,
		},
	}

	operations, err := list[*metadata.ExportToYT, metadata.ExportToYT](
		ctx, c.operationServiceClient, &request.ListOperationsRequest,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return &operations.listOperations, nil
}

// ListExecuteQuery returns list of query executions
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func (c *Client) ListExecuteQuery(ctx context.Context, opts ...options.List) (
	*listOperationsWithNextToken[*metadata.ExecuteQuery, metadata.ExecuteQuery], error,
) {
	request := &options.ListOperationsRequest{
		ListOperationsRequest: Ydb_Operations.ListOperationsRequest{
			Kind: kindExecuteQuery,
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(request)
		}
	}

	operations, err := list[*metadata.ExecuteQuery, metadata.ExecuteQuery](
		ctx, c.operationServiceClient, &request.ListOperationsRequest,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return operations, nil
}

func New(ctx context.Context, balancer grpc.ClientConnInterface) *Client {
	return &Client{
		operationServiceClient: Ydb_Operation_V1.NewOperationServiceClient(
			conn.WithContextModifier(balancer, conn.WithoutWrapping),
		),
	}
}
