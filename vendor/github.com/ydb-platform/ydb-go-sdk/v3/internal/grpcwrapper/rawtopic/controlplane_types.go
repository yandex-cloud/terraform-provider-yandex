package rawtopic

import (
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawoptional"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	errUnexpectedNilPartitioningSettings     = xerrors.Wrap(errors.New("ydb: unexpected nil partitioning settings"))
	errUnexpecredNilAutoPartitioningSettings = xerrors.Wrap(errors.New("ydb: unexpected nil auto-partitioning settings"))
	errUnexpectedNilAutoPartitionWriteSpeed  = xerrors.Wrap(errors.New("ydb: unexpected nil auto-partition write speed"))
)

type Consumer struct {
	Name            string
	Important       bool
	SupportedCodecs rawtopiccommon.SupportedCodecs
	ReadFrom        rawoptional.Time
	Attributes      map[string]string
}

func (c *Consumer) MustFromProto(consumer *Ydb_Topic.Consumer) {
	c.Name = consumer.GetName()
	c.Important = consumer.GetImportant()
	c.Attributes = consumer.GetAttributes()
	c.ReadFrom.MustFromProto(consumer.GetReadFrom())
	c.SupportedCodecs.MustFromProto(consumer.GetSupportedCodecs())
}

func (c *Consumer) ToProto() *Ydb_Topic.Consumer {
	return &Ydb_Topic.Consumer{
		Name:            c.Name,
		Important:       c.Important,
		ReadFrom:        c.ReadFrom.ToProto(),
		SupportedCodecs: c.SupportedCodecs.ToProto(),
		Attributes:      c.Attributes,
	}
}

type MeteringMode int

const (
	MeteringModeUnspecified      = MeteringMode(Ydb_Topic.MeteringMode_METERING_MODE_UNSPECIFIED)
	MeteringModeReservedCapacity = MeteringMode(Ydb_Topic.MeteringMode_METERING_MODE_RESERVED_CAPACITY)
	MeteringModeRequestUnits     = MeteringMode(Ydb_Topic.MeteringMode_METERING_MODE_REQUEST_UNITS)
)

type PartitioningSettings struct {
	MinActivePartitions      int64
	MaxActivePartitions      int64
	PartitionCountLimit      int64
	AutoPartitioningSettings AutoPartitioningSettings
}

func (s *PartitioningSettings) FromProto(proto *Ydb_Topic.PartitioningSettings) error {
	if proto == nil {
		return xerrors.WithStackTrace(errUnexpectedNilPartitioningSettings)
	}

	s.MinActivePartitions = proto.GetMinActivePartitions()
	s.MaxActivePartitions = proto.GetMaxActivePartitions()
	s.PartitionCountLimit = proto.GetPartitionCountLimit() //nolint:staticcheck

	return s.AutoPartitioningSettings.FromProto(proto.GetAutoPartitioningSettings())
}

func (s *PartitioningSettings) ToProto() *Ydb_Topic.PartitioningSettings {
	return &Ydb_Topic.PartitioningSettings{
		MinActivePartitions:      s.MinActivePartitions,
		MaxActivePartitions:      s.MaxActivePartitions,
		PartitionCountLimit:      s.PartitionCountLimit,
		AutoPartitioningSettings: s.AutoPartitioningSettings.ToProto(),
	}
}

type AutoPartitioningSettings struct {
	AutoPartitioningStrategy           AutoPartitioningStrategy
	AutoPartitioningWriteSpeedStrategy AutoPartitioningWriteSpeedStrategy
}

func (s *AutoPartitioningSettings) ToProto() *Ydb_Topic.AutoPartitioningSettings {
	if s == nil {
		return nil
	}

	return &Ydb_Topic.AutoPartitioningSettings{
		Strategy:            s.AutoPartitioningStrategy.ToProto(),
		PartitionWriteSpeed: s.AutoPartitioningWriteSpeedStrategy.ToProto(),
	}
}

func (s *AutoPartitioningSettings) FromProto(proto *Ydb_Topic.AutoPartitioningSettings) error {
	if proto == nil {
		return xerrors.WithStackTrace(errUnexpecredNilAutoPartitioningSettings)
	}
	s.AutoPartitioningStrategy = AutoPartitioningStrategy(proto.GetStrategy())

	if proto.GetPartitionWriteSpeed() != nil {
		if err := s.AutoPartitioningWriteSpeedStrategy.FromProto(proto.GetPartitionWriteSpeed()); err != nil {
			return err
		}
	}

	return nil
}

type AutoPartitioningStrategy int32

const (
	AutoPartitioningStrategyUnspecified    = AutoPartitioningStrategy(Ydb_Topic.AutoPartitioningStrategy_AUTO_PARTITIONING_STRATEGY_UNSPECIFIED)       //nolint:lll
	AutoPartitioningStrategyDisabled       = AutoPartitioningStrategy(Ydb_Topic.AutoPartitioningStrategy_AUTO_PARTITIONING_STRATEGY_DISABLED)          //nolint:lll
	AutoPartitioningStrategyScaleUp        = AutoPartitioningStrategy(Ydb_Topic.AutoPartitioningStrategy_AUTO_PARTITIONING_STRATEGY_SCALE_UP)          //nolint:lll
	AutoPartitioningStrategyScaleUpAndDown = AutoPartitioningStrategy(Ydb_Topic.AutoPartitioningStrategy_AUTO_PARTITIONING_STRATEGY_SCALE_UP_AND_DOWN) //nolint:lll
	AutoPartitioningStrategyPaused         = AutoPartitioningStrategy(Ydb_Topic.AutoPartitioningStrategy_AUTO_PARTITIONING_STRATEGY_PAUSED)            //nolint:lll
)

func (s AutoPartitioningStrategy) ToProto() Ydb_Topic.AutoPartitioningStrategy {
	return Ydb_Topic.AutoPartitioningStrategy(s)
}

type AutoPartitioningWriteSpeedStrategy struct {
	StabilizationWindow    rawoptional.Duration
	UpUtilizationPercent   int32
	DownUtilizationPercent int32
}

func (s *AutoPartitioningWriteSpeedStrategy) ToProto() *Ydb_Topic.AutoPartitioningWriteSpeedStrategy {
	return &Ydb_Topic.AutoPartitioningWriteSpeedStrategy{
		StabilizationWindow:    s.StabilizationWindow.ToProto(),
		UpUtilizationPercent:   s.UpUtilizationPercent,
		DownUtilizationPercent: s.DownUtilizationPercent,
	}
}

func (s *AutoPartitioningWriteSpeedStrategy) FromProto(speed *Ydb_Topic.AutoPartitioningWriteSpeedStrategy) error {
	if speed == nil {
		return xerrors.WithStackTrace(errUnexpectedNilAutoPartitionWriteSpeed)
	}

	s.StabilizationWindow.MustFromProto(speed.GetStabilizationWindow())
	s.UpUtilizationPercent = speed.GetUpUtilizationPercent()
	s.DownUtilizationPercent = speed.GetDownUtilizationPercent()

	return nil
}

type AlterPartitioningSettings struct {
	SetMinActivePartitions        rawoptional.Int64
	SetMaxActivePartitions        rawoptional.Int64
	SetPartitionCountLimit        rawoptional.Int64
	AlterAutoPartitioningSettings *AlterAutoPartitioningSettings
}

func (s *AlterPartitioningSettings) ToProto() *Ydb_Topic.AlterPartitioningSettings {
	return &Ydb_Topic.AlterPartitioningSettings{
		SetMinActivePartitions:        s.SetMinActivePartitions.ToProto(),
		SetMaxActivePartitions:        s.SetMaxActivePartitions.ToProto(),
		SetPartitionCountLimit:        s.SetPartitionCountLimit.ToProto(),
		AlterAutoPartitioningSettings: s.AlterAutoPartitioningSettings.ToProto(),
	}
}

type AlterAutoPartitioningSettings struct {
	SetStrategy            AutoPartitioningStrategy
	SetPartitionWriteSpeed *AlterAutoPartitioningWriteSpeedStrategy
}

func (s *AlterAutoPartitioningSettings) ToProto() *Ydb_Topic.AlterAutoPartitioningSettings {
	if s == nil {
		return nil
	}
	strategy := s.SetStrategy.ToProto()

	return &Ydb_Topic.AlterAutoPartitioningSettings{
		SetStrategy:            &strategy,
		SetPartitionWriteSpeed: s.SetPartitionWriteSpeed.ToProto(),
	}
}

type AlterAutoPartitioningWriteSpeedStrategy struct {
	SetStabilizationWindow    rawoptional.Duration
	SetUpUtilizationPercent   rawoptional.Int32
	SetDownUtilizationPercent rawoptional.Int32
}

func (s *AlterAutoPartitioningWriteSpeedStrategy) ToProto() *Ydb_Topic.AlterAutoPartitioningWriteSpeedStrategy {
	if s == nil {
		return nil
	}

	return &Ydb_Topic.AlterAutoPartitioningWriteSpeedStrategy{
		SetStabilizationWindow:    s.SetStabilizationWindow.ToProto(),
		SetUpUtilizationPercent:   s.SetUpUtilizationPercent.ToProto(),
		SetDownUtilizationPercent: s.SetDownUtilizationPercent.ToProto(),
	}
}
