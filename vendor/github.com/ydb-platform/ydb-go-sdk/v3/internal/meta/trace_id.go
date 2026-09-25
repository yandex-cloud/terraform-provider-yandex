package meta

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type newTraceIDOpts struct {
	newRandom func() (uuid.UUID, error)
}

func TraceID(ctx context.Context, opts ...func(opts *newTraceIDOpts)) (context.Context, string, error) {
	if id, has := traceID(ctx); has {
		return ctx, id, nil
	}
	options := newTraceIDOpts{
		newRandom: uuid.NewRandom,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	uuid, err := options.newRandom()
	if err != nil {
		return ctx, "", xerrors.WithStackTrace(err)
	}
	id := uuid.String()

	return metadata.AppendToOutgoingContext(ctx, HeaderTraceID, id), id, nil
}
