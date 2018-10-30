// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package operation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-sdk/sdkerrors"
)

type Client = operation.OperationServiceClient
type Proto = operation.Operation

func New(client Client, proto *Proto) *Operation {
	if proto == nil {
		panic("nil operation")
	}
	return &Operation{proto: proto, client: client, newTimer: defaultTimer}
}

func defaultTimer(d time.Duration) (func() <-chan time.Time, func() bool) {
	timer := time.NewTimer(d)
	return func() <-chan time.Time {
		return timer.C
	}, timer.Stop
}

type Operation struct {
	proto    *Proto
	client   Client
	newTimer func(time.Duration) (func() <-chan time.Time, func() bool)
}

func (o *Operation) Proto() *Proto  { return o.proto }
func (o *Operation) Client() Client { return o.client }

//revive:disable:var-naming
func (o *Operation) Id() string { return o.proto.Id }

//revive:enable:var-naming
func (o *Operation) Description() string { return o.proto.Description }
func (o *Operation) CreatedBy() string   { return o.proto.CreatedBy }

func (o *Operation) CreatedAt() time.Time {
	ts, err := ptypes.Timestamp(o.proto.CreatedAt)
	if err != nil {
		panic(fmt.Sprintf("invalid created at: %v", err))
	}
	return ts
}

func (o *Operation) Metadata() (proto.Message, error) {
	return UnmarshalAny(o.RawMetadata())
}

func (o *Operation) RawMetadata() *any.Any { return o.proto.Metadata }

func (o *Operation) Error() error {
	st := o.ErrorStatus()
	if st == nil {
		return nil
	}
	return st.Err()
}

func (o *Operation) ErrorStatus() *status.Status {
	proto := o.proto.GetError()
	if proto == nil {
		return nil
	}
	return status.FromProto(proto)
}

func (o *Operation) Response() (proto.Message, error) {
	resp := o.RawResponse()
	if resp == nil {
		return nil, nil
	}
	return UnmarshalAny(resp)
}

func (o *Operation) RawResponse() *any.Any {
	return o.proto.GetResponse()
}

func (o *Operation) Done() bool   { return o.proto.Done }
func (o *Operation) Ok() bool     { return o.Done() && o.proto.GetResponse() != nil }
func (o *Operation) Failed() bool { return o.Done() && o.proto.GetError() != nil }

// Poll gets new state of operation from operation client. On success operation state is updated.
// Returns error if state get failed.
func (o *Operation) Poll(ctx context.Context, opts ...grpc.CallOption) error {
	req := &operation.GetOperationRequest{OperationId: o.Id()}
	state, err := o.Client().Get(ctx, req, opts...)
	if err != nil {
		return err
	}
	o.proto = state
	return nil
}

// Cancel requests operation cancel. On success operation state is updated.
// Returns error if cancel failed.
func (o *Operation) Cancel(ctx context.Context, opts ...grpc.CallOption) error {
	req := &operation.CancelOperationRequest{OperationId: o.Id()}
	state, err := o.Client().Cancel(ctx, req, opts...)
	if err != nil {
		return err
	}
	o.proto = state
	return nil
}

func (o *Operation) Wait(ctx context.Context, opts ...grpc.CallOption) error {
	return o.WaitInterval(ctx, DefaultPollInterval, opts...)
}

const (
	pollIntervalMetadataKey = "x-operation-poll-interval"
)

func (o *Operation) waitInterval(ctx context.Context, pollInterval time.Duration, opts ...grpc.CallOption) error {
	var headers metadata.MD
	opts = append(opts, grpc.Header(&headers))

	// https://st.yandex-team.ru/MCDEV-860
	const maxNotFoundRetry = 3
	notFoundCount := 0
	for !o.Done() {
		headers = metadata.MD{}
		err := o.Poll(ctx, opts...)
		if err != nil {
			if notFoundCount < maxNotFoundRetry && shoudRetry(err) {
				notFoundCount++
			} else {
				// Message needed to distinguish poll fail and operation error, which are both gRPC status.
				return sdkerrors.WithMessage(err, "poll fail")
			}
		}
		if o.Done() {
			break
		}
		interval := pollInterval
		if vals := headers.Get(pollIntervalMetadataKey); len(vals) > 0 {
			i, err := strconv.Atoi(vals[0])
			if err == nil {
				interval = time.Duration(i) * time.Second
			}
		}
		if interval <= 0 {
			continue
		}
		wait, stop := o.newTimer(interval)
		select {
		case <-wait():
		case <-ctx.Done():
			stop()
			return ctx.Err()
		}
	}
	return o.Error()
}

func shoudRetry(err error) bool {
	status, ok := status.FromError(err)
	return ok && status.Code() == codes.NotFound
}

func (o *Operation) WaitInterval(ctx context.Context, pollInterval time.Duration, opts ...grpc.CallOption) error {
	return sdkerrors.WithOperationID(o.waitInterval(ctx, pollInterval, opts...), o.Id())
}

const DefaultPollInterval = time.Second

// TODO(skipor): per operation call options needed?
type Operations []*Operation

// Batch is helper to combine multiple operations.
// Example:
// err := Batch(operation1, operation2).Wait(ctx, callOpt1, callOpt2)
func Batch(ops ...*Operation) Operations {
	return Operations(ops)
}

func (ops Operations) Wait(ctx context.Context, opts ...grpc.CallOption) error {
	return ops.WaitInterval(ctx, DefaultPollInterval, opts...)
}

func (ops Operations) WaitInterval(ctx context.Context, poll time.Duration, opts ...grpc.CallOption) error {
	var errs error
	for _, op := range ops {
		err := op.WaitInterval(ctx, poll, opts...)
		if err == nil {
			continue
		}
		if ctx.Err() != nil {
			// Wait is canceled, errs don't matter anymore.
			return ctx.Err()
		}
		errs = sdkerrors.Append(errs, err)
	}
	return errs
}

func (ops Operations) Cancel(ctx context.Context, opts ...grpc.CallOption) error {
	var errs error
	for _, op := range ops {
		err := op.Cancel(ctx, opts...)
		if err == nil {
			continue
		}
		if ctx.Err() != nil {
			// Cancel is canceled, errs don't matter anymore.
			return ctx.Err()
		}
		err = sdkerrors.WithOperationID(err, op.Id())
		errs = sdkerrors.Append(errs, err)
	}
	return errs
}
