package topictypes

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/clone"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawoptional"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopiccommon"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topiclistenerinternal"
)

// Codec code for use in topics
// Allow to use custom values in interval [10000,20000)
type Codec int32

const (
	CodecRaw  = Codec(rawtopiccommon.CodecRaw)
	CodecGzip = Codec(rawtopiccommon.CodecGzip)

	// CodecLzop not supported by default, customer need provide own codec library
	CodecLzop = Codec(rawtopiccommon.CodecLzop)

	// CodecZstd not supported by default, customer need provide own codec library
	CodecZstd = Codec(rawtopiccommon.CodecZstd)

	CodecCustomerFirst = Codec(rawtopiccommon.CodecCustomerFirst)
	CodecCustomerEnd   = Codec(rawtopiccommon.CodecCustomerEnd) // last allowed custom codec id is CodecCustomerEnd-1
)

func (c Codec) ToRaw(r *rawtopiccommon.Codec) {
	*r = rawtopiccommon.Codec(c)
}

// Consumer contains info about topic consumer
type Consumer struct {
	Name            string
	Important       bool
	SupportedCodecs []Codec
	ReadFrom        time.Time
	Attributes      map[string]string
}

// ToRaw public format to internal. Used internally only.
func (c *Consumer) ToRaw(raw *rawtopic.Consumer) {
	raw.Name = c.Name
	raw.Important = c.Important
	raw.Attributes = c.Attributes

	raw.SupportedCodecs = make(rawtopiccommon.SupportedCodecs, len(c.SupportedCodecs))
	for index, codec := range c.SupportedCodecs {
		raw.SupportedCodecs[index] = rawtopiccommon.Codec(codec)
	}

	if !c.ReadFrom.IsZero() {
		raw.ReadFrom.HasValue = true
		raw.ReadFrom.Value = c.ReadFrom
	}
	raw.Attributes = c.Attributes
}

// FromRaw convert internal format to public. Used internally only.
func (c *Consumer) FromRaw(raw *rawtopic.Consumer) {
	c.Attributes = raw.Attributes
	c.Important = raw.Important
	c.Name = raw.Name

	c.SupportedCodecs = make([]Codec, len(raw.SupportedCodecs))
	for index, codec := range raw.SupportedCodecs {
		c.SupportedCodecs[index] = Codec(codec)
	}

	if raw.ReadFrom.HasValue {
		c.ReadFrom = raw.ReadFrom.Value
	}
}

// MeteringMode mode of topic's metering. Used for serverless installations.
type MeteringMode int

const (
	MeteringModeUnspecified      = MeteringMode(rawtopic.MeteringModeUnspecified)
	MeteringModeReservedCapacity = MeteringMode(rawtopic.MeteringModeReservedCapacity)
	MeteringModeRequestUnits     = MeteringMode(rawtopic.MeteringModeRequestUnits)
)

// FromRaw convert from internal format to public. Used internally only.
func (m *MeteringMode) FromRaw(raw rawtopic.MeteringMode) {
	*m = MeteringMode(raw)
}

// ToRaw convert from public format to internal. Used internally only.
func (m *MeteringMode) ToRaw(raw *rawtopic.MeteringMode) {
	*raw = rawtopic.MeteringMode(*m)
}

// PartitionSettings settings of partitions
type PartitionSettings struct {
	MinActivePartitions      int64
	MaxActivePartitions      int64
	PartitionCountLimit      int64
	AutoPartitioningSettings AutoPartitioningSettings
}

// ToRaw convert public format to internal. Used internally only.
func (s *PartitionSettings) ToRaw(raw *rawtopic.PartitioningSettings) {
	raw.MinActivePartitions = s.MinActivePartitions
	raw.MaxActivePartitions = s.MaxActivePartitions
	raw.PartitionCountLimit = s.PartitionCountLimit
	s.AutoPartitioningSettings.ToRaw(&raw.AutoPartitioningSettings)
}

// FromRaw convert internal format to public. Used internally only.
func (s *PartitionSettings) FromRaw(raw *rawtopic.PartitioningSettings) {
	s.MinActivePartitions = raw.MinActivePartitions
	s.MaxActivePartitions = raw.MaxActivePartitions
	s.PartitionCountLimit = raw.PartitionCountLimit
	s.AutoPartitioningSettings.FromRaw(&raw.AutoPartitioningSettings)
}

// AutoPartitioningSettings contains settings for automatic partitioning
type AutoPartitioningSettings struct {
	AutoPartitioningStrategy           AutoPartitioningStrategy
	AutoPartitioningWriteSpeedStrategy AutoPartitioningWriteSpeedStrategy
}

// ToRaw convert public format to internal. Used internally only.
func (s *AutoPartitioningSettings) ToRaw(raw *rawtopic.AutoPartitioningSettings) {
	raw.AutoPartitioningStrategy = rawtopic.AutoPartitioningStrategy(s.AutoPartitioningStrategy)
	raw.AutoPartitioningWriteSpeedStrategy = s.AutoPartitioningWriteSpeedStrategy.ToRaw()
}

// FromRaw convert internal format to public. Used internally only.
func (s *AutoPartitioningSettings) FromRaw(raw *rawtopic.AutoPartitioningSettings) {
	s.AutoPartitioningStrategy = AutoPartitioningStrategy(raw.AutoPartitioningStrategy)
	s.AutoPartitioningWriteSpeedStrategy.FromRaw(&raw.AutoPartitioningWriteSpeedStrategy)
}

// AutoPartitioningStrategy defines the strategy for automatic partitioning
type AutoPartitioningStrategy int32

const (
	AutoPartitioningStrategyUnspecified    = AutoPartitioningStrategy(rawtopic.AutoPartitioningStrategyUnspecified)
	AutoPartitioningStrategyDisabled       = AutoPartitioningStrategy(rawtopic.AutoPartitioningStrategyDisabled)
	AutoPartitioningStrategyScaleUp        = AutoPartitioningStrategy(rawtopic.AutoPartitioningStrategyScaleUp)
	AutoPartitioningStrategyScaleUpAndDown = AutoPartitioningStrategy(rawtopic.AutoPartitioningStrategyScaleUpAndDown)
	AutoPartitioningStrategyPaused         = AutoPartitioningStrategy(rawtopic.AutoPartitioningStrategyPaused)
)

// AutoPartitioningWriteSpeedStrategy contains settings for write speed strategy
type AutoPartitioningWriteSpeedStrategy struct {
	StabilizationWindow    time.Duration
	UpUtilizationPercent   int32
	DownUtilizationPercent int32
}

// ToRaw convert public format to internal. Used internally only.
func (s *AutoPartitioningWriteSpeedStrategy) ToRaw() rawtopic.AutoPartitioningWriteSpeedStrategy {
	return rawtopic.AutoPartitioningWriteSpeedStrategy{
		StabilizationWindow: rawoptional.Duration{
			Value:    s.StabilizationWindow,
			HasValue: s.StabilizationWindow > 0,
		},
		UpUtilizationPercent:   s.UpUtilizationPercent,
		DownUtilizationPercent: s.DownUtilizationPercent,
	}
}

// FromRaw convert internal format to public. Used internally only.
func (s *AutoPartitioningWriteSpeedStrategy) FromRaw(raw *rawtopic.AutoPartitioningWriteSpeedStrategy) {
	s.StabilizationWindow = raw.StabilizationWindow.ToDurationWithDefault()
	s.UpUtilizationPercent = raw.UpUtilizationPercent
	s.DownUtilizationPercent = raw.DownUtilizationPercent
}

// TopicDescription contains info about topic.
type TopicDescription struct {
	Path                              string
	PartitionSettings                 PartitionSettings
	Partitions                        []PartitionInfo
	RetentionPeriod                   time.Duration
	RetentionStorageMB                int64
	SupportedCodecs                   []Codec
	PartitionWriteBurstBytes          int64
	PartitionWriteSpeedBytesPerSecond int64
	Attributes                        map[string]string
	Consumers                         []Consumer
	MeteringMode                      MeteringMode
}

// FromRaw convert from public format to internal. Used internally only.
func (d *TopicDescription) FromRaw(raw *rawtopic.DescribeTopicResult) {
	d.Path = raw.Self.Name
	d.PartitionSettings.FromRaw(&raw.PartitioningSettings)

	d.Partitions = make([]PartitionInfo, len(raw.Partitions))
	for i := range raw.Partitions {
		d.Partitions[i].FromRaw(&raw.Partitions[i])
	}

	d.RetentionPeriod = raw.RetentionPeriod
	d.RetentionStorageMB = raw.RetentionStorageMB

	d.SupportedCodecs = make([]Codec, len(raw.SupportedCodecs))
	for i := 0; i < len(raw.SupportedCodecs); i++ {
		d.SupportedCodecs[i] = Codec(raw.SupportedCodecs[i])
	}

	d.PartitionWriteSpeedBytesPerSecond = raw.PartitionWriteSpeedBytesPerSecond
	d.PartitionWriteBurstBytes = raw.PartitionWriteBurstBytes

	d.RetentionPeriod = raw.RetentionPeriod
	d.RetentionStorageMB = raw.RetentionStorageMB

	d.Attributes = make(map[string]string)
	for k, v := range raw.Attributes {
		d.Attributes[k] = v
	}

	d.Consumers = make([]Consumer, len(raw.Consumers))
	for i := 0; i < len(raw.Consumers); i++ {
		d.Consumers[i].FromRaw(&raw.Consumers[i])
	}

	d.MeteringMode.FromRaw(raw.MeteringMode)
}

// PartitionInfo contains info about partition.
type PartitionInfo struct {
	PartitionID        int64
	Active             bool
	ChildPartitionIDs  []int64
	ParentPartitionIDs []int64
	PartitionStats     PartitionStats
}

// FromRaw convert from internal format to public. Used internally only.
func (p *PartitionInfo) FromRaw(raw *rawtopic.PartitionInfo) {
	p.PartitionID = raw.PartitionID
	p.Active = raw.Active

	p.ChildPartitionIDs = clone.Int64Slice(raw.ChildPartitionIDs)
	p.ParentPartitionIDs = clone.Int64Slice(raw.ParentPartitionIDs)
	p.PartitionStats.FromRaw(&raw.PartitionStats)
}

type MultipleWindowsStat struct {
	PerMinute int64
	PerHour   int64
	PerDay    int64
}

func (m *MultipleWindowsStat) FromRaw(raw *rawtopic.MultipleWindowsStat) {
	if raw != nil {
		m.PerMinute = raw.PerMinute
		m.PerHour = raw.PerHour
		m.PerDay = raw.PerDay
	}
}

type OffsetRange topiclistenerinternal.PublicOffsetsRange

type PartitionStats struct {
	PartitionsOffset OffsetRange
	StoreSizeBytes   int64
	LastWriteTime    *time.Time
	MaxWriteTimeLag  *time.Duration
	BytesWritten     MultipleWindowsStat
}

func (p *PartitionStats) FromRaw(raw *rawtopic.PartitionStats) {
	p.PartitionsOffset.Start = raw.PartitionsOffset.Start.ToInt64()
	p.PartitionsOffset.End = raw.PartitionsOffset.End.ToInt64()
	p.StoreSizeBytes = raw.StoreSizeBytes
	p.LastWriteTime = raw.LastWriteTime.ToTime()
	p.MaxWriteTimeLag = raw.MaxWriteTimeLag.ToDuration()
	p.BytesWritten.FromRaw(&raw.BytesWritten)
}

type PartitionConsumerStats struct {
	LastReadOffset                 int64
	CommittedOffset                int64
	ReadSessionID                  string
	PartitionReadSessionCreateTime *time.Time
	LastReadTime                   *time.Time
	MaxReadTimeLag                 *time.Duration
	MaxWriteTimeLag                *time.Duration
	BytesRead                      MultipleWindowsStat
	ReaderName                     string
}

func (s *PartitionConsumerStats) FromRaw(raw *rawtopic.PartitionConsumerStats) {
	s.LastReadOffset = raw.LastReadOffset
	s.CommittedOffset = raw.CommittedOffset
	s.ReadSessionID = raw.ReadSessionID
	s.PartitionReadSessionCreateTime = raw.PartitionReadSessionCreateTime.ToTime()
	s.LastReadTime = raw.LastReadTime.ToTime()
	s.MaxReadTimeLag = raw.MaxReadTimeLag.ToDuration()
	s.MaxWriteTimeLag = raw.MaxWriteTimeLag.ToDuration()
	s.BytesRead.FromRaw(&raw.BytesRead)
	s.ReaderName = raw.ReaderName
}

type DescribeConsumerPartitionInfo struct {
	PartitionID            int64
	Active                 bool
	ChildPartitionIDs      []int64
	ParentPartitionIDs     []int64
	PartitionStats         PartitionStats
	PartitionConsumerStats PartitionConsumerStats
}

func (p *DescribeConsumerPartitionInfo) FromRaw(raw *rawtopic.DescribeConsumerResultPartitionInfo) {
	p.PartitionID = raw.PartitionID
	p.Active = raw.Active
	p.ChildPartitionIDs = clone.Int64Slice(raw.ChildPartitionIDs)
	p.ParentPartitionIDs = clone.Int64Slice(raw.ParentPartitionIDs)
	p.PartitionStats.FromRaw(&raw.PartitionStats)
	p.PartitionConsumerStats.FromRaw(&raw.PartitionConsumerStats)
}

type TopicConsumerDescription struct {
	Path       string
	Consumer   Consumer
	Partitions []DescribeConsumerPartitionInfo
}

func (d *TopicConsumerDescription) FromRaw(raw *rawtopic.DescribeConsumerResult) {
	d.Path = raw.Self.Name
	d.Consumer.FromRaw(&raw.Consumer)

	d.Partitions = make([]DescribeConsumerPartitionInfo, len(raw.Partitions))
	for i := range raw.Partitions {
		d.Partitions[i].FromRaw(&raw.Partitions[i])
	}
}
