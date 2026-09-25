package operation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// pollIntervalMetadataKey is a metadata key used to retrieve the polling interval value from gRPC response headers.
const pollIntervalMetadataKey = "x-operation-poll-interval"

// Error extracts and returns the error from the provided YCOperation if it exists, otherwise returns nil.
func Error(op YCOperation) error {
	pb := op.GetError()
	if pb == nil {
		return nil
	}

	st := status.FromProto(pb)
	if st == nil {
		return nil
	}

	return st.Err()
}

// Response processes the response from a YCOperation and unmarshals it into a proto.Message.
func Response(op YCOperation) (proto.Message, error) {
	raw := op.GetResponse()
	if raw == nil {
		return nil, nil
	}

	return raw.UnmarshalNew()
}

// PollFunc defines a function type used to poll the status of an operation using its ID and return a YCOperation or an error.
type PollFunc func(ctx context.Context, operationId string, opts ...grpc.CallOption) (YCOperation, error)

// waitInterval polls the operation until its completion or context timeout, using a custom polling interval strategy.
func waitInterval(
	ctx context.Context,
	operationId string,
	poll PollFunc,
	pollInterval PollIntervalFunc,
	opts ...grpc.CallOption,
) (YCOperation, error) {
	op, err := PollUntilDone(ctx, operationId, poll, pollInterval, opts...)
	if err != nil {
		return nil, err
	}

	err = Error(op)
	if err != nil {
		return nil, fmt.Errorf("operation (id=%s) failed: %w", op.GetId(), err)
	}

	return op, nil
}

// PollUntilDone repeatedly polls the status of a long-running operation until it is marked as done or the context is canceled.
// ctx is the context for managing the polling lifecycle and cancellation.
// operationId identifies the specific operation to be polled.
// poll is a function used to fetch the current state of the operation.
// pollInterval determines the duration to wait between polling attempts based on the attempt count.
// opts allows for additional gRPC call options to be passed during polling.
// Returns the completed YCOperation or an error if polling fails or the context is canceled.
func PollUntilDone(
	ctx context.Context,
	operationId string,
	poll PollFunc,
	pollInterval PollIntervalFunc,
	opts ...grpc.CallOption,
) (YCOperation, error) {
	var headers metadata.MD
	opts = append(opts, grpc.Header(&headers))

	// Sometimes, the returned operation is not on all replicas yet,
	// so we need to ignore first couple of NotFound errors.
	const maxNotFoundRetry = 3

	notFoundCount := 0
	attempt := 0

	for {
		headers = metadata.MD{}

		polledOperation, err := poll(ctx, operationId, opts...)
		if err != nil {
			st, ok := status.FromError(err)
			if notFoundCount < maxNotFoundRetry && ok && st.Code() == codes.NotFound {
				notFoundCount++
			} else {
				// Message needed to distinguish poll fail and operation error, which are both gRPC status.
				return nil, fmt.Errorf("operation (id=%s) poll fail: %w", operationId, err)
			}
		} else {
			if polledOperation.GetDone() {
				return polledOperation, nil
			}
		}

		interval := pollInterval(attempt)
		attempt++

		if values := headers.Get(pollIntervalMetadataKey); len(values) > 0 {
			i, err := strconv.Atoi(values[0])
			if err == nil {
				interval = time.Duration(i) * time.Second
			}
		}

		if interval <= 0 {
			continue
		}

		waitTimer := time.NewTimer(interval)
		select {
		case <-waitTimer.C:
		case <-ctx.Done():
			waitTimer.Stop()
			return nil, fmt.Errorf("operation (id=%s) wait context done: %w", operationId, ctx.Err())
		}
	}
}
