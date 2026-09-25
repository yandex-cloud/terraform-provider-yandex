package rawydb

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

type StatusCode int

const (
	StatusSuccess       = StatusCode(Ydb.StatusIds_SUCCESS)
	StatusInternalError = StatusCode(Ydb.StatusIds_INTERNAL_ERROR)
)

func (s *StatusCode) FromProto(p Ydb.StatusIds_StatusCode) error {
	*s = StatusCode(p)

	return nil
}

func (s StatusCode) IsSuccess() bool {
	return s == StatusSuccess
}

func (s StatusCode) String() string {
	return Ydb.StatusIds_StatusCode(s).String()
}

// Equals compares this StatusCode with another StatusCode for equality
func (s StatusCode) Equals(other StatusCode) bool {
	return s == other
}
