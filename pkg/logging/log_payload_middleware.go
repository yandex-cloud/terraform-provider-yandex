package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HeaderLoggingDecider is a user-provided function for deciding whether to log the header with a given key
// request/response payloads
type HeaderLoggingDecider func(key string) bool

func defaultLogPayloadOptions() *logPayloadOptions {
	return &logPayloadOptions{
		marshaller: DefaultJSONPBMarshal,
		header:     func(string) bool { return true },
	}
}

func statusFromError(err error) (s *status.Status, ok bool) {
	if err == nil {
		return status.New(codes.OK, ""), true
	}
	if se, ok := err.(interface {
		GRPCStatus() *status.Status
	}); ok {
		return se.GRPCStatus(), true
	}
	return status.New(codes.Internal, err.Error()), false
}

func NewLogPayloadMiddleware(options ...LogPayloadClientOption) grpc.UnaryClientInterceptor {
	opts := defaultLogPayloadOptions()
	for _, v := range options {
		v(opts)
	}
	middleware := &logPayloadClientMiddleware{
		helper: logHelper{opts},
	}
	return middleware.InterceptUnary
}

type logPayloadClientMiddleware struct {
	helper logHelper
	header HeaderLoggingDecider //nolint:structcheck,megacheck
}

func (m *logPayloadClientMiddleware) InterceptUnary(
	ctx context.Context,
	method string,
	req, resp interface{},
	conn *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var header metadata.MD
	var trailer metadata.MD
	opts = append(opts, grpc.Header(&header), grpc.Trailer(&trailer))
	md, _ := metadata.FromOutgoingContext(ctx)
	m.helper.request(ctx, reqEntry{
		method:  method,
		message: req,
		md:      md,
	})
	err := invoker(ctx, method, req, resp, conn, opts...)
	m.helper.response(ctx, respEntry{
		method:  method,
		message: resp,
		err:     err,
		header:  header,
		trailer: trailer,
	})
	return err
}

type logHelper struct {
	options *logPayloadOptions
}

type reqEntry struct {
	method  string
	message interface{}
	md      metadata.MD
}

func (h *logHelper) request(ctx context.Context, ent reqEntry) {
	// Extra space after 'Request' makes log entry aligned with 'Response' entry :)
	h.log(ctx, "request", "Request ", logEntry{
		method:  ent.method,
		message: ent.message,
		header:  ent.md,
	})
}

type respEntry struct {
	method          string
	message         interface{}
	err             error
	header, trailer metadata.MD
}

func (h *logHelper) response(ctx context.Context, ent respEntry) {
	h.log(ctx, "response", "Response", logEntry(ent))
}

func (h *logHelper) filterMeta(md metadata.MD) metadata.MD {
	x := make(metadata.MD, len(md))
	for k, v := range md {
		if !h.options.header(k) {
			v = []string{"*** hidden ***"}
		}
		x[k] = v
	}
	return x
}

func (h *logHelper) log(ctx context.Context, key string, msg string, ent logEntry) {
	var payload interface{}

	if !IsNil(ent.message) {
		p, ok := ent.message.(proto.Message)
		if ok {
			payload = &jsonpbMarshaller{p, h.options.marshaller}
		} else {
			// Payload should be nil or proto.Message, but let's not panic if not - just log it as JSON.
			payload = ent.message
		}
	}
	// if err is gRPC error, marshall it as proto.Message
	var outErr interface{} = ent.err
	var statusCode string
	if ent.err != nil {
		st, ok := statusFromError(ent.err)
		pb := st.Proto()
		if ok && pb != nil {
			statusCode = codeString(st.Code())
			outErr = &jsonpbMarshaller{pb, h.options.marshaller}
		}
	}
	bytes, err := json.Marshal(jsonMessage{
		Method:     ent.method,
		Header:     h.filterMeta(ent.header),
		Trailer:    h.filterMeta(ent.trailer),
		Payload:    payload,
		StatusCode: statusCode,
		Error:      outErr,
	})
	if err == nil {
		log.Print("[DEBUG] ", getLogMessage(msg, ent.method), string(bytes))
	} else {
		log.Print("[DEBUG] Failed to marshal json message", err)
	}
}

// getLogMessage appends short method name to message for better log readability.
// Because when the full method name appears in the middle of log fields it is difficult to read it on a small display.
// "Request",  "/yandex.cloud.priv.compute.v1.InstanceService/Delete" -> "Request InstanceService/Delete"
func getLogMessage(baseMsg string, method string) string {
	packageDotIndex := strings.LastIndexByte(method, '.')
	if packageDotIndex < 0 {
		// Empty package or unexpected trash instead of method name, ignore.
		return baseMsg
	}
	serviceNameIndex := packageDotIndex + 1
	if serviceNameIndex >= len(method) {
		// Trash instead of method name, ignore.
		return baseMsg
	}
	shortMethod := method[serviceNameIndex:]
	return baseMsg + " " + shortMethod
}

type logEntry struct {
	method          string
	message         interface{}
	err             error
	header, trailer metadata.MD
}

type jsonMessage struct {
	Method     string      `json:"method"`
	Header     metadata.MD `json:"header,omitempty"`
	Trailer    metadata.MD `json:"trailer,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	StatusCode string      `json:"status_code,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

type JSONPBMarshaller func(m proto.Message) ([]byte, error)

var defaultJSONPBMarshaller = &jsonpb.Marshaler{OrigName: true}

func DefaultJSONPBMarshal(m proto.Message) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := defaultJSONPBMarshaller.Marshal(b, m); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

type jsonpbMarshaller struct {
	proto.Message
	JSONPBMarshaller
}

func (j *jsonpbMarshaller) MarshalJSON() ([]byte, error) {
	return j.JSONPBMarshaller(j.Message)
}

// codeString returns a string by code in upper snake case as this is the way to name these codes in gRPC standard
func codeString(c codes.Code) string {
	// code.String() returns string in UpperCamelCase, so it shouldn't be used
	return codeToStr[c]
}

var codeToStr = map[codes.Code]string{
	codes.OK:                 "OK",
	codes.Canceled:           "CANCELLED",
	codes.Unknown:            "UNKNOWN",
	codes.InvalidArgument:    "INVALID_ARGUMENT",
	codes.DeadlineExceeded:   "DEADLINE_EXCEEDED",
	codes.NotFound:           "NOT_FOUND",
	codes.AlreadyExists:      "ALREADY_EXISTS",
	codes.PermissionDenied:   "PERMISSION_DENIED",
	codes.ResourceExhausted:  "RESOURCE_EXHAUSTED",
	codes.FailedPrecondition: "FAILED_PRECONDITION",
	codes.Aborted:            "ABORTED",
	codes.OutOfRange:         "OUT_OF_RANGE",
	codes.Unimplemented:      "UNIMPLEMENTED",
	codes.Internal:           "INTERNAL",
	codes.Unavailable:        "UNAVAILABLE",
	codes.DataLoss:           "DATA_LOSS",
	codes.Unauthenticated:    "UNAUTHENTICATED",
}
