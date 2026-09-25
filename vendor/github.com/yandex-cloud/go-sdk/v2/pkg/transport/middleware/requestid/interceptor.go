package requestid

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	requestIDHeader     = "x-request-id"
	clientTraceIDHeader = "x-client-trace-id"
	serverTraceIDHeader = "x-server-trace-id"
)

// Interceptor is a gRPC unary client interceptor for propagating trace and request IDs across service boundaries.
// It injects client-side headers and captures server-side headers to track the request lifecycle.
// The interceptor appends metadata and wraps errors with correlated request and trace information.
func Interceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		clientTraceID := uuid.New().String()
		requestID := uuid.New().String()

		md, ok := metadata.FromOutgoingContext(ctx)

		if ok && len(md.Get(clientTraceIDHeader)) > 0 {
			clientTraceID = md.Get(clientTraceIDHeader)[0]
		}
		if ok && len(md.Get(requestIDHeader)) > 0 {
			requestID = md.Get(requestIDHeader)[0]
		}

		ctx = withMetadata(ctx, map[string]string{
			requestIDHeader:     requestID,
			clientTraceIDHeader: clientTraceID,
		})

		var responseHeader, responseTrailer metadata.MD
		opts = append(opts, grpc.Header(&responseHeader), grpc.Trailer(&responseTrailer))
		err := invoker(ctx, method, req, reply, conn, opts...)

		return wrapError(err, clientTraceID, requestID, responseHeader, responseTrailer)
	}
}

// RequestIDs represents a collection of identifiers for tracking client and server request and trace information.
// ClientTraceID is a unique identifier for tracing client requests.
// RequestID is a unique identifier for server requests, typically derived from server response headers.
// ServerTraceID is a unique trace identifier for server-side operations, typically derived from server response headers.
type RequestIDs struct {
	ClientTraceID string
	RequestID     string
	ServerTraceID string
}

// ErrorWithRequestIDs wraps an original error and associates it with client and server request/trace IDs.
type ErrorWithRequestIDs struct {
	OrigErr error
	IDs     RequestIDs
}

// Error returns a formatted error message including server and client request/trace IDs, followed by the original error message.
func (e *ErrorWithRequestIDs) Error() (msg string) {
	if e.IDs.RequestID != "" {
		msg += fmt.Sprintf("%s = %s ", requestIDHeader, e.IDs.RequestID)
	}

	if e.IDs.ServerTraceID != "" {
		msg += fmt.Sprintf("%s = %s ", serverTraceIDHeader, e.IDs.ServerTraceID)
	}

	if e.IDs.ClientTraceID != "" {
		msg += fmt.Sprintf("%s = %s ", clientTraceIDHeader, e.IDs.ClientTraceID)
	}

	return msg + e.OrigErr.Error()
}

// GRPCStatus converts the original error into a gRPC status representation and returns the associated status.
func (e ErrorWithRequestIDs) GRPCStatus() *status.Status {
	return status.Convert(e.OrigErr)
}

// Unwrap returns the original error for use with errors.Is and errors.As.
func (e *ErrorWithRequestIDs) Unwrap() error {
	return e.OrigErr
}

// RequestIDsFromError extracts RequestIDs from an error if it contains request-related context. Returns ok as true if successful.
// Returns a copy of RequestIDs to prevent mutation of the error's internal state.
func RequestIDsFromError(err error) (*RequestIDs, bool) {
	if withID, ok := err.(*ErrorWithRequestIDs); ok {
		ids := withID.IDs
		return &ids, true
	}

	return nil, false
}

// ContextWithClientTraceID returns a new context embedding the provided clientTraceID for outgoing requests.
func ContextWithClientTraceID(ctx context.Context, clientTraceID string) context.Context {
	return withMetadata(ctx, map[string]string{
		clientTraceIDHeader: clientTraceID,
	})
}

// wrapError wraps an error with client and server request/trace IDs, enriching error context with additional metadata.
// If the error is already wrapped with request IDs, it returns the original error. Returns nil if the input error is nil.
func wrapError(err error, clientTraceID, requestID string, responseHeader, responseTrailer metadata.MD) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*ErrorWithRequestIDs); ok {
		return err
	}

	serverRequestID := getKeyFromMD(responseHeader, requestIDHeader)
	if serverRequestID == "" {
		serverRequestID = getKeyFromMD(responseTrailer, requestIDHeader)
	}
	if serverRequestID == "" {
		serverRequestID = requestID
	}

	serverTraceID := getKeyFromMD(responseHeader, serverTraceIDHeader)
	if serverTraceID == "" {
		serverTraceID = getKeyFromMD(responseTrailer, serverTraceIDHeader)
	}

	return &ErrorWithRequestIDs{
		err,
		RequestIDs{
			ClientTraceID: clientTraceID,
			RequestID:     serverRequestID,
			ServerTraceID: serverTraceID,
		},
	}
}

// getKeyFromMD retrieves the first value associated with the given key from the metadata.MD headers.
// Returns an empty string if the key is not present.
func getKeyFromMD(md metadata.MD, key string) string {
	keyValues := md.Get(key)
	if len(keyValues) == 0 {
		return ""
	}

	return keyValues[0]
}

// withMetadata adds the specified metadata to the outgoing gRPC context and returns the updated context.
func withMetadata(ctx context.Context, meta map[string]string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}

	for k, v := range meta {
		md.Set(k, v)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
