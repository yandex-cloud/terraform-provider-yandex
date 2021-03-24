package logging

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

var marshaller = &jsonpb.Marshaler{OrigName: true}

func HideSensitiveValues(m proto.Message) proto.Message {
	if IsNil(m) {
		return nil
	}
	m = proto.Clone(m)
	HideSensitive(m)
	return m
}

type WithHideSensitive interface {
	HideSensitive()
}

// HideSensitive hides the sensitive fields
func HideSensitive(m proto.Message) bool {
	if m == nil {
		return false
	}
	if _, ok := m.(WithHideSensitive); ok {
		m.(WithHideSensitive).HideSensitive()
		return true
	}
	return false
}

func JSONHidingSensitiveValuesMarshaller(m proto.Message) ([]byte, error) {
	m = HideSensitiveValues(m)
	b := &bytes.Buffer{}
	if err := marshaller.Marshal(b, m); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func IsNil(x interface{}) bool {
	return x == nil || reflect.ValueOf(x).IsNil()
}

func HeaderIsNotSensitive(key string) bool {
	if key == ":authority" {
		return true
	}
	return !strings.Contains(key, "auth") && !strings.Contains(key, "token")
}

func NewAPILoggingUnaryInterceptor() grpc.UnaryClientInterceptor {
	return grpc_middleware.ChainUnaryClient(NewLogPayloadMiddleware(
		LogPayloadClientMarshaller(JSONHidingSensitiveValuesMarshaller),
		LogPayloadClientHeader(HeaderIsNotSensitive),
	))
}
