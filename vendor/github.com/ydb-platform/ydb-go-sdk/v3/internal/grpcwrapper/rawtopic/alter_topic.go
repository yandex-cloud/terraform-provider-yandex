package rawtopic

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawoptional"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

type AlterTopicRequest struct {
	OperationParams rawydb.OperationParams

	Path string

	AlterPartitionSettings               AlterPartitioningSettings
	SetRetentionPeriod                   rawoptional.Duration
	SetRetentionStorageMB                rawoptional.Int64
	SetSupportedCodecs                   bool
	SetSupportedCodecsValue              rawtopiccommon.SupportedCodecs
	SetPartitionWriteSpeedBytesPerSecond rawoptional.Int64
	SetPartitionWriteBurstBytes          rawoptional.Int64
	AlterAttributes                      map[string]string
	AddConsumers                         []Consumer
	DropConsumers                        []string
	AlterConsumers                       []AlterConsumer
	SetMeteringMode                      MeteringMode
}

func (req *AlterTopicRequest) ToProto() *Ydb_Topic.AlterTopicRequest {
	res := &Ydb_Topic.AlterTopicRequest{
		OperationParams:                      req.OperationParams.ToProto(),
		Path:                                 req.Path,
		AlterPartitioningSettings:            req.AlterPartitionSettings.ToProto(),
		SetRetentionPeriod:                   req.SetRetentionPeriod.ToProto(),
		SetRetentionStorageMb:                req.SetRetentionStorageMB.ToProto(),
		SetPartitionWriteSpeedBytesPerSecond: req.SetPartitionWriteSpeedBytesPerSecond.ToProto(),
		SetPartitionWriteBurstBytes:          req.SetPartitionWriteBurstBytes.ToProto(),
		AlterAttributes:                      req.AlterAttributes,
		SetMeteringMode:                      Ydb_Topic.MeteringMode(req.SetMeteringMode),
	}

	if req.SetSupportedCodecs {
		res.SetSupportedCodecs = req.SetSupportedCodecsValue.ToProto()
	}

	res.AddConsumers = make([]*Ydb_Topic.Consumer, len(req.AddConsumers))
	for i := range req.AddConsumers {
		res.AddConsumers[i] = req.AddConsumers[i].ToProto()
	}

	res.DropConsumers = req.DropConsumers

	res.AlterConsumers = make([]*Ydb_Topic.AlterConsumer, len(req.AlterConsumers))
	for i := range req.AlterConsumers {
		res.AlterConsumers[i] = req.AlterConsumers[i].ToProto()
	}

	return res
}

type AlterTopicResult struct {
	Operation rawydb.Operation
}

func (r *AlterTopicResult) FromProto(proto *Ydb_Topic.AlterTopicResponse) error {
	return r.Operation.FromProtoWithStatusCheck(proto.GetOperation())
}

type AlterConsumer struct {
	Name               string
	SetImportant       rawoptional.Bool
	SetReadFrom        rawoptional.Time
	SetSupportedCodecs rawtopiccommon.SupportedCodecs
	AlterAttributes    map[string]string
}

func (c *AlterConsumer) ToProto() *Ydb_Topic.AlterConsumer {
	return &Ydb_Topic.AlterConsumer{
		Name:               c.Name,
		SetImportant:       c.SetImportant.ToProto(),
		SetReadFrom:        c.SetReadFrom.ToProto(),
		SetSupportedCodecs: c.SetSupportedCodecs.ToProto(),
		AlterAttributes:    c.AlterAttributes,
	}
}
