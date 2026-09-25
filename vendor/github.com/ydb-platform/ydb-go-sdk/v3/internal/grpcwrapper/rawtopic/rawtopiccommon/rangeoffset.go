package rawtopiccommon

import (
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var errUnexpectedProtobufInOffsets = xerrors.Wrap(errors.New("ydb: unexpected protobuf nil offsets"))

type OffsetRange struct {
	Start Offset
	End   Offset
}

func (r *OffsetRange) FromProto(p *Ydb_Topic.OffsetsRange) error {
	if p == nil {
		return xerrors.WithStackTrace(errUnexpectedProtobufInOffsets)
	}

	r.Start.FromInt64(p.GetStart())
	r.End.FromInt64(p.GetEnd())

	return nil
}

func (r *OffsetRange) ToProto() *Ydb_Topic.OffsetsRange {
	return &Ydb_Topic.OffsetsRange{
		Start: r.Start.ToInt64(),
		End:   r.End.ToInt64(),
	}
}
