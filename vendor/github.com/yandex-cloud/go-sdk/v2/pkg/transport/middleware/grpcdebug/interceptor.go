// Package grpcdebug provides gRPC client interceptors that log every outgoing
// unary call and stream message at zap Debug level. Intended to be installed
// only when the user explicitly opts in via --debug-grpc; the cost of proto
// marshalling is gated by the logger's level so the interceptor is a no-op
// when its logger is not configured to emit Debug.
//
// Output format mirrors the legacy ycp (syntax 1) shape so that log analysis
// tools and humans see the same envelope for both syntaxes:
//
//	Request  IamTokenService/Create  {"request_id": "...", "request": {"method":"/...","header":{...},"payload":{...}}}
//	Response IamTokenService/Create  {"request_id": "...", "response": {"method":"/...","trailer":{...},"payload":{},"status_code":"UNAUTHENTICATED","error":{...}}}
package grpcdebug

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// hiddenPlaceholder is the literal value substituted for sensitive proto
// payload fields (matches legacy ycp syntax-1 output).
const hiddenPlaceholder = "*** hidden ***"

// sensitivePayloadFields lists proto field names (snake_case) whose values must
// be replaced with hiddenPlaceholder before logging, no matter how the value
// looks. Walked recursively over the JSON tree of every payload.
var sensitivePayloadFields = map[string]struct{}{
	"yandex_passport_oauth_token": {},
	"jwt":                         {},
	"iam_token":                   {},
	"subject_token":               {},
	"actor_token":                 {},
	"refresh_token":               {},
	"access_token":                {},
	"private_key":                 {},
	"secret":                      {},
	"password":                    {},
}

// sensitiveHeaderNames lists outgoing gRPC metadata keys that we drop entirely
// from the logged "header" / "trailer" maps. Their values are credentials we
// must never write to a trace file.
var sensitiveHeaderNames = map[string]struct{}{
	"authorization":     {},
	"x-ydb-auth-ticket": {},
}

func enabled(logger *zap.Logger) bool {
	return logger != nil && logger.Core().Enabled(zapcore.DebugLevel)
}

// shortMethod turns "/yandex.cloud.priv.iam.v1.IamTokenService/Create" into
// "IamTokenService/Create" — the bit users actually scan for.
func shortMethod(method string) string {
	trimmed := strings.TrimPrefix(method, "/")
	slash := strings.LastIndex(trimmed, "/")
	if slash < 0 {
		return trimmed
	}
	left := trimmed[:slash]
	if dot := strings.LastIndex(left, "."); dot >= 0 {
		return left[dot+1:] + trimmed[slash:]
	}
	return trimmed
}

// filterHeaders returns a copy of md with sensitive entries dropped.
func filterHeaders(md metadata.MD) map[string][]string {
	if len(md) == 0 {
		return nil
	}
	out := make(map[string][]string, len(md))
	for k, v := range md {
		if _, sensitive := sensitiveHeaderNames[strings.ToLower(k)]; sensitive {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// outgoingHeaders extracts metadata that the client is about to send for this
// call. Includes only context-stored values (x-request-id, x-trace-id,
// idempotency-key, …); the per-call authorization Bearer header is filtered.
func outgoingHeaders(ctx context.Context) map[string][]string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil
	}
	return filterHeaders(md)
}

// extractRequestID pulls the first non-empty x-request-id from the metadata
// map. Used to surface a top-level "request_id" field, matching legacy ycp.
func extractRequestID(headers map[string][]string) string {
	for _, key := range []string{"x-request-id", "X-Request-Id", "X-Request-ID"} {
		if v, ok := headers[key]; ok && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

// requestEnvelope / responseEnvelope are the structured logging shapes — they
// serialize via zap.Any to the JSON form expected for parity with syntax 1.
type requestEnvelope struct {
	Method  string              `json:"method"`
	Header  map[string][]string `json:"header,omitempty"`
	Payload interface{}         `json:"payload,omitempty"`
}

type responseEnvelope struct {
	Method     string              `json:"method"`
	Trailer    map[string][]string `json:"trailer,omitempty"`
	Payload    interface{}         `json:"payload,omitempty"`
	StatusCode string              `json:"status_code,omitempty"`
	Error      *responseError      `json:"error,omitempty"`
}

type responseError struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

// buildResponseError converts an invoker error into the structured form. Nil
// for the success path.
func buildResponseError(err error) (string, *responseError) {
	if err == nil {
		return "", nil
	}
	st, _ := status.FromError(err)
	if st == nil {
		return codes.Unknown.String(), &responseError{Code: uint32(codes.Unknown), Message: err.Error()}
	}
	return strings.ToUpper(st.Code().String()), &responseError{
		Code:    uint32(st.Code()),
		Message: st.Message(),
	}
}

// UnaryClientInterceptor logs every outgoing unary RPC: a "Request" entry
// before invoke and a "Response" entry after. Both lines carry the structured
// envelope (method / header / trailer / payload / status_code / error) used by
// the legacy ycp syntax-1 logger.
func UnaryClientInterceptor(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !enabled(logger) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		short := shortMethod(method)
		header := outgoingHeaders(ctx)
		reqID := extractRequestID(header)

		reqPayload := protoMessagePayload(req)
		logger.Debug("Request "+short,
			zap.String("request_id", reqID),
			zap.Any("request", requestEnvelope{
				Method:  method,
				Header:  header,
				Payload: reqPayload,
			}),
		)

		var respHeader, respTrailer metadata.MD
		callOpts := append([]grpc.CallOption{grpc.Header(&respHeader), grpc.Trailer(&respTrailer)}, opts...)
		err := invoker(ctx, method, req, reply, cc, callOpts...)

		statusCode, respErr := buildResponseError(err)
		trailer := filterHeaders(respTrailer)
		var respPayload interface{}
		if err == nil {
			respPayload = protoMessagePayload(reply)
		}
		logger.Debug("Response "+short,
			zap.String("request_id", reqID),
			zap.Any("response", responseEnvelope{
				Method:     method,
				Trailer:    trailer,
				Payload:    respPayload,
				StatusCode: statusCode,
				Error:      respErr,
			}),
		)
		return err
	}
}

// StreamClientInterceptor logs the open of every stream plus each message sent
// or received. Payloads are masked using the same rules as the unary path.
func StreamClientInterceptor(logger *zap.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !enabled(logger) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		short := shortMethod(method)
		header := outgoingHeaders(ctx)
		reqID := extractRequestID(header)

		logger.Debug("Stream open "+short,
			zap.String("request_id", reqID),
			zap.Any("request", requestEnvelope{
				Method: method,
				Header: header,
			}),
			zap.Bool("client_streams", desc.ClientStreams),
			zap.Bool("server_streams", desc.ServerStreams),
		)

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			statusCode, respErr := buildResponseError(err)
			logger.Debug("Stream open failed "+short,
				zap.String("request_id", reqID),
				zap.Any("response", responseEnvelope{
					Method:     method,
					StatusCode: statusCode,
					Error:      respErr,
				}),
			)
			return nil, err
		}

		return &loggingClientStream{
			ClientStream: stream,
			logger:       logger,
			method:       method,
			short:        short,
			requestID:    reqID,
		}, nil
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger    *zap.Logger
	method    string
	short     string
	requestID string
}

func (s *loggingClientStream) SendMsg(m interface{}) error {
	s.logger.Debug("Stream send "+s.short,
		zap.String("request_id", s.requestID),
		zap.Any("request", requestEnvelope{
			Method:  s.method,
			Payload: protoMessagePayload(m),
		}),
	)
	return s.ClientStream.SendMsg(m)
}

func (s *loggingClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		statusCode, respErr := buildResponseError(err)
		s.logger.Debug("Stream recv "+s.short,
			zap.String("request_id", s.requestID),
			zap.Any("response", responseEnvelope{
				Method:     s.method,
				StatusCode: statusCode,
				Error:      respErr,
			}),
		)
		return err
	}
	s.logger.Debug("Stream recv "+s.short,
		zap.String("request_id", s.requestID),
		zap.Any("response", responseEnvelope{
			Method:  s.method,
			Payload: protoMessagePayload(m),
		}),
	)
	return nil
}

// protoMessagePayload converts msg into a JSON-ish payload (map / slice /
// scalar) and masks sensitive fields. Returns nil if the value is not a proto
// message — this keeps the structured log entry minimal in that edge case.
func protoMessagePayload(m interface{}) interface{} {
	pm, ok := m.(proto.Message)
	if !ok || pm == nil {
		return nil
	}
	return cleanProtoPayload(pm)
}
