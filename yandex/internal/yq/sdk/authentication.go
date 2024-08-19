package sdk

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type yqMDMiddleware struct {
	authToken string
	folderID  string
}

func newYQMDMiddleware(authToken, folderID string) *yqMDMiddleware {
	return &yqMDMiddleware{
		authToken: authToken,
		folderID:  folderID,
	}
}

func (c *yqMDMiddleware) unaryInterceptor(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = c.contextWithMetadata(ctx)
	return invoker(ctx, method, req, reply, conn, opts...)
}

func (c *yqMDMiddleware) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = c.contextWithMetadata(ctx)
	return streamer(ctx, desc, conn, method, opts...)
}

func (c *yqMDMiddleware) contextWithMetadata(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		"x-ydb-fq-project":  fmt.Sprintf("yandexcloud://%s", c.folderID),
		"x-ydb-auth-ticket": c.authToken,
	})

	return metadata.NewOutgoingContext(ctx, md)
}
