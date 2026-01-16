package utils

import (
	"testing"

	"github.com/c2h5oh/datasize"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func AssertProtoEqual(t *testing.T, testname string, expected, actual proto.Message) {
	t.Helper()

	if proto.Equal(expected, actual) {
		return
	}

	t.Errorf("Unexpected request in test %s", testname)

	expJSON, _ := protojson.MarshalOptions{
		Multiline:       true,
		EmitUnpopulated: true,
	}.Marshal(expected)

	actJSON, _ := protojson.MarshalOptions{
		Multiline:       true,
		EmitUnpopulated: true,
	}.Marshal(actual)

	t.Logf("expected JSON:\n%s", expJSON)
	t.Logf("actual JSON:\n%s", actJSON)
}

func ToGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}
