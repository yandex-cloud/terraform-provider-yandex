package yqsdk

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type yqMDMiddleware struct {
	authTokenProvider authTokenProviderType
	folderID          string
}

func newYQMDMiddleware(authTokenProvider authTokenProviderType, folderID string) *yqMDMiddleware {
	return &yqMDMiddleware{
		authTokenProvider: authTokenProvider,
		folderID:          folderID,
	}
}

func (c *yqMDMiddleware) unaryInterceptor(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, err := c.contextWithMetadata(ctx)
	if err != nil {
		return err
	}
	return invoker(ctx, method, req, reply, conn, opts...)
}

func (c *yqMDMiddleware) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx, err := c.contextWithMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return streamer(ctx, desc, conn, method, opts...)
}

func (c *yqMDMiddleware) contextWithMetadata(ctx context.Context) (context.Context, error) {
	token, err := c.authTokenProvider(ctx)
	if err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"x-ydb-fq-project":  fmt.Sprintf("yandexcloud://%s", c.folderID),
		"x-ydb-auth-ticket": token,
	})

	return metadata.NewOutgoingContext(ctx, md), nil
}
