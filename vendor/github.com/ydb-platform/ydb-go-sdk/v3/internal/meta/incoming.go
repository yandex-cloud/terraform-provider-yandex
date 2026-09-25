package meta

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type (
	MetadataCallback    func(md metadata.MD)
	metadataCallbackKey struct{}
)

func WithTrailerCallback(ctx context.Context, callback MetadataCallback) context.Context {
	if existingCallback, has := ctx.Value(metadataCallbackKey{}).(MetadataCallback); has {
		return context.WithValue(ctx, metadataCallbackKey{}, MetadataCallback(
			func(md metadata.MD) {
				existingCallback(md)
				callback(md)
			},
		))
	}

	return context.WithValue(ctx, metadataCallbackKey{}, callback)
}

func CallTrailerCallback(ctx context.Context, md metadata.MD) {
	if len(md) == 0 {
		return
	}
	callback, has := ctx.Value(metadataCallbackKey{}).(MetadataCallback)
	if !has {
		return
	}
	callback(md)
}
