package yandex

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	clientTraceIDHeader   = "x-client-trace-id"
	clientRequestIDHeader = "x-client-request-id"
	serverRequestIDHeader = "x-request-id"
	serverTraceIDHeader   = "x-server-trace-id"
)

type requestIDs struct {
	ClientTraceID   string
	ClientRequestID string
	ServerRequestID string
	ServerTraceID   string
}

type errorWithRequestIDs struct {
	origErr error
	ids     requestIDs
}

func (e *errorWithRequestIDs) Error() string {
	switch {
	case e.ids.ServerRequestID != "":
		return fmt.Sprintf("request-id = %s %s", e.ids.ServerRequestID, e.origErr.Error())
	case e.ids.ClientRequestID != "":
		return fmt.Sprintf("client-request-id = %s %s", e.ids.ClientRequestID, e.origErr.Error())
	default:
		return e.origErr.Error()
	}
}
func (e errorWithRequestIDs) GRPCStatus() *status.Status {
	return status.Convert(e.origErr)
}

func requestIDInterceptor() func(ctx context.Context, method string, req interface{}, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	clientTraceID := uuid.New().String()
	return func(ctx context.Context, method string, req interface{}, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		clientRequestID := uuid.New().String()
		var responseHeader metadata.MD
		opts = append(opts, grpc.Header(&responseHeader))
		ctx = addClientRequestIDs(ctx, clientTraceID, clientRequestID)
		err := invoker(ctx, method, req, reply, conn, opts...)
		return WrapError(err, clientTraceID, clientRequestID, responseHeader)
	}
}

func WrapError(err error, clientTraceID, clientRequestID string, responseHeader metadata.MD) error {
	if err == nil {
		return nil
	}
	if _, ok := requestIDsFromError(err); ok {
		return err
	}

	serverRequestID := getServerHeader(responseHeader, serverRequestIDHeader)
	serverTraceID := getServerHeader(responseHeader, serverTraceIDHeader)

	return &errorWithRequestIDs{
		err,
		requestIDs{
			ClientTraceID:   clientTraceID,
			ClientRequestID: clientRequestID,
			ServerRequestID: serverRequestID,
			ServerTraceID:   serverTraceID,
		},
	}
}

func requestIDsFromError(err error) (*requestIDs, bool) {
	if withID, ok := err.(*errorWithRequestIDs); ok {
		return &withID.ids, ok
	}
	return nil, false
}

func getServerHeader(responseHeader metadata.MD, key string) string {
	serverHeaderIDRaw := responseHeader.Get(key)
	if len(serverHeaderIDRaw) == 0 {
		return ""
	}

	return serverHeaderIDRaw[0]
}

func addClientRequestIDs(ctx context.Context, clientTraceID, clientRequestID string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	md.Set(clientRequestIDHeader, clientRequestID)
	md.Set(clientTraceIDHeader, clientTraceID)
	return metadata.NewOutgoingContext(ctx, md)
}
