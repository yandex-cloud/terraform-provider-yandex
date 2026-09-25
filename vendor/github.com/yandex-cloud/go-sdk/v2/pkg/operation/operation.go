package operation

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ AbstractOperation = (*Operation)(nil)

// Operation represents an operation instance with associated metadata, response, and error handling functionality.
type Operation struct {
	proto          YCOperation
	concretization *Concretization

	metadata      proto.Message
	response      proto.Message
	responseError error
}

// Concretization is a type that defines the operation handling behavior for polling and response management.
// Poll is a function for retrieving operation information from an operation service.
// MetadataType specifies the protobuf message type for the operation's metadata.
// ResponseType specifies the protobuf message type for the operation's response.
// GetResourceID is a function to extract a resource ID from the metadata.
type Concretization struct {
	Poll PollFunc

	MetadataType proto.Message
	ResponseType proto.Message

	GetResourceID func(metadata proto.Message) string
}

// PollIntervalFunc defines a function type that calculates the delay between polling attempts based on the attempt number.
type PollIntervalFunc func(attempt int) time.Duration

// NewOperation creates a new Operation instance based on the provided YCOperation and Concretization parameters.
// Returns an Operation instance and an error if the metadata type does not match or other issues arise.
func NewOperation(pb YCOperation, concretization *Concretization) (*Operation, error) {
	var data proto.Message
	var err error

	if pb.GetMetadata() != nil {
		data, err = pb.GetMetadata().UnmarshalNew()

		if err != nil {
			return nil, err
		}

	}

	if reflect.TypeOf(data) != reflect.TypeOf(concretization.MetadataType) {
		return nil, fmt.Errorf("expected operation metadata to be '%s', but got '%s'",
			proto.MessageName(concretization.MetadataType),
			proto.MessageName(data),
		)
	}

	op := &Operation{
		proto:          pb,
		concretization: concretization,
		metadata:       data,
		response:       nil,
		responseError:  nil,
	}

	if op.Done() {
		resp, err := op.parseResponse()
		if err != nil {
			return nil, err
		}

		// Ignore this error, we can return the operation with filled `responseError`
		_ = op.fillResponse(resp)
	}

	return op, nil
}

// Abstract returns the current operation as an AbstractOperation, providing access to its abstract interface.
func (o *Operation) Abstract() AbstractOperation {
	return o
}

// ID returns the identifier of the Operation as a string.
func (o *Operation) ID() string { return o.proto.GetId() }

// Description returns a string that provides the description of the operation based on the underlying proto definition.
func (o *Operation) Description() string { return o.proto.GetDescription() }

// CreatedBy returns the identifier of the entity that created the operation.
func (o *Operation) CreatedBy() string { return o.proto.GetCreatedBy() }

// CreatedAt returns the creation timestamp of the operation as a time.Time object.
func (o *Operation) CreatedAt() time.Time {
	return o.proto.GetCreatedAt().AsTime()
}

// Metadata retrieves the metadata associated with the Operation. It returns a proto.Message representing the metadata.
func (o *Operation) Metadata() proto.Message {
	return o.metadata
}

// ResourceID retrieves the resource ID associated with the operation's metadata or panics if not defined.
func (o *Operation) ResourceID() string {
	if o.concretization.GetResourceID == nil {
		panic(fmt.Errorf("this operation's metadata does not have a resource id"))
	}

	return o.concretization.GetResourceID(o.metadata)
}

// Done checks if the operation has been completed and returns true if it is done, otherwise false.
func (o *Operation) Done() bool { return o.proto.GetDone() }

// PollOnce performs a single polling attempt to update the operation's state and metadata in place. Returns an error if polling fails.
func (o *Operation) PollOnce(ctx context.Context, opts ...grpc.CallOption) error {
	pb, err := o.concretization.Poll(ctx, o.proto.GetId(), opts...)
	if err != nil {
		return err
	}

	next, err := NewOperation(pb, o.concretization)
	if err != nil {
		return err
	}

	*o = *next

	return nil
}

// defaultPollIntervalFunc returns the default polling interval as a constant duration of one second.
func defaultPollIntervalFunc(int) time.Duration {
	return time.Second
}

// Wait blocks until the operation is completed or the context is canceled, returning the operation's response or an error.
func (o *Operation) Wait(ctx context.Context, opts ...grpc.CallOption) (proto.Message, error) {
	return o.WaitInterval(ctx, defaultPollIntervalFunc, opts...)
}

// WaitInterval polls the operation periodically until it is complete or the context is canceled, using a custom interval.
func (o *Operation) WaitInterval(
	ctx context.Context,
	pollInterval PollIntervalFunc,
	opts ...grpc.CallOption,
) (proto.Message, error) {
	if o.Done() {
		if o.responseError != nil {
			return nil, o.responseError
		}

		if err := Error(o.proto); err != nil {
			return nil, fmt.Errorf("operation (id=%s) failed: %w", o.ID(), err)
		}

		return o.response, nil
	}

	op, err := waitInterval(ctx, o.proto.GetId(), o.concretization.Poll, pollInterval, opts...)
	if err != nil {
		return nil, err
	}

	next, err := NewOperation(op, o.concretization)
	if err != nil {
		return nil, err
	}

	*o = *next

	if o.responseError != nil {
		return nil, o.responseError
	}

	return o.response, nil
}

// Error returns the error encountered during the operation, prioritizing the responseError if available.
func (o *Operation) Error() error {
	if o.responseError != nil {
		return o.responseError
	}

	return Error(o.proto)
}

// Response returns the response of the completed operation. Panics if the operation is not completed or has errors.
func (o *Operation) Response() proto.Message {
	if !o.Done() {
		panic("getting response from a not completed operation")
	}

	if o.responseError != nil {
		// This error was returned from Wait, and should has been handled by Wait caller.
		panic(o.responseError)
	}

	return o.response
}

// parseResponse extracts and unmarshals the response from an operation after it is completed. Returns nil if no response exists.
func (o *Operation) parseResponse() (proto.Message, error) {
	if !o.Done() {
		panic("parsing response from a not completed operation")
	}

	raw := o.proto.GetResponse()
	if raw == nil {
		return nil, nil
	}

	return raw.UnmarshalNew()
}

// fillResponse sets the operation's response and validates its type against the expected ResponseType.
// Returns an error if the provided response type does not match the expected type.
// If the expected ResponseType is *emptypb.Empty, it initializes the response as such.
func (o *Operation) fillResponse(response proto.Message) error {
	if reflect.TypeOf(o.concretization.ResponseType) == reflect.TypeOf((*emptypb.Empty)(nil)) {
		response = &emptypb.Empty{}
	} else if reflect.TypeOf(response) != reflect.TypeOf(o.concretization.ResponseType) {
		o.responseError = fmt.Errorf("expected operation response to be '%s', but got '%s'",
			proto.MessageName(o.concretization.ResponseType), proto.MessageName(response))
		return o.responseError
	}

	o.response = response

	return nil
}

// Result returns the result of the operation if it has completed. Panics if the operation is not yet done.
func (o *Operation) Result() OperationResult {
	if !o.Done() {
		panic("getting result from a not completed operation")
	}

	return o
}
