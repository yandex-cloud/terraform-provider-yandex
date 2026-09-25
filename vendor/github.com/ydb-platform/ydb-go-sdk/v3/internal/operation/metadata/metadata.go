package metadata

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Export"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Import"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/options"
)

type (
	TypesConstraint interface {
		BuildIndex | ImportFromS3 | ExportToS3 | ExportToYT | ExecuteQuery
	}
	Constraint[T TypesConstraint] interface {
		fromProto(metadata *anypb.Any) *T
	}
	BuildIndex struct {
		Description string
		State       string
		Progress    float32
	}
	ImportFromS3 struct {
		Settings string
		Status   string
		Items    []string
	}
	ExportToS3 struct {
		Settings string
		Status   string
		Items    []string
	}
	ExportToYT struct {
		Settings string
		Status   string
		Items    []string
	}
	ExecuteQuery options.MetadataExecuteQuery
)

func (*ImportFromS3) fromProto(metadata *anypb.Any) *ImportFromS3 { //nolint:unused
	var pb Ydb_Import.ImportFromS3Metadata
	if err := metadata.UnmarshalTo(&pb); err != nil {
		panic(err)
	}

	return &ImportFromS3{
		Settings: pb.GetSettings().String(),
		Status:   pb.GetProgress().String(),
		Items: func() (items []string) {
			for _, item := range pb.GetItemsProgress() {
				items = append(items, item.String())
			}

			return items
		}(),
	}
}

func (*ExportToS3) fromProto(metadata *anypb.Any) *ExportToS3 { //nolint:unused
	var pb Ydb_Export.ExportToS3Metadata
	if err := metadata.UnmarshalTo(&pb); err != nil {
		panic(err)
	}

	return &ExportToS3{
		Settings: pb.GetSettings().String(),
		Status:   pb.GetProgress().String(),
		Items: func() (items []string) {
			for _, item := range pb.GetItemsProgress() {
				items = append(items, item.String())
			}

			return items
		}(),
	}
}

func (*ExportToYT) fromProto(metadata *anypb.Any) *ExportToYT { //nolint:unused
	var pb Ydb_Export.ExportToYtMetadata
	if err := metadata.UnmarshalTo(&pb); err != nil {
		panic(err)
	}

	return &ExportToYT{
		Settings: pb.GetSettings().String(),
		Status:   pb.GetProgress().String(),
		Items: func() (items []string) {
			for _, item := range pb.GetItemsProgress() {
				items = append(items, item.String())
			}

			return items
		}(),
	}
}

func (*ExecuteQuery) fromProto(metadata *anypb.Any) *ExecuteQuery { //nolint:unused
	return (*ExecuteQuery)(options.ToMetadataExecuteQuery(metadata))
}

func (*BuildIndex) fromProto(metadata *anypb.Any) *BuildIndex { //nolint:unused
	var pb Ydb_Table.IndexBuildMetadata
	if err := metadata.UnmarshalTo(&pb); err != nil {
		panic(err)
	}

	return &BuildIndex{
		Description: pb.GetDescription().String(),
		State:       pb.GetState().String(),
		Progress:    pb.GetProgress(),
	}
}

func FromProto[PT Constraint[T], T TypesConstraint](metadata *anypb.Any) *T {
	var pt PT

	return pt.fromProto(metadata)
}
