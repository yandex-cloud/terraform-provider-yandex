package validate

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// isStatusWithCode checks if any nested error matches provided code
func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	if check {
		return true
	}

	if nestedErr := errors.Unwrap(err); nestedErr != nil {
		return IsStatusWithCode(nestedErr, code)
	}

	return check
}

func ProtoDump(msg proto.Message) string {
	data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("**ERROR json mashal failed: %+v** message dump: %s", err, spewConfig.Sdump(msg))
	}
	return string(data)
}

var spewConfig = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
}
