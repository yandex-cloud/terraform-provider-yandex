package options

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/feature"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type Column struct {
	Name   string
	Type   types.Type
	Family string
}

func (c Column) toYDB() *Ydb_Table.ColumnMeta {
	return &Ydb_Table.ColumnMeta{
		Name:   c.Name,
		Type:   types.TypeToYDB(c.Type),
		Family: c.Family,
	}
}

func NewTableColumn(name string, typ types.Type, family string) Column {
	return Column{
		Name:   name,
		Type:   typ,
		Family: family,
	}
}

type IndexDescription struct {
	Name         string
	IndexColumns []string
	DataColumns  []string
	Status       Ydb_Table.TableIndexDescription_Status
	Type         IndexType
}

type Description struct {
	Name                 string
	Columns              []Column
	PrimaryKey           []string
	KeyRanges            []KeyRange
	Stats                *TableStats
	ColumnFamilies       []ColumnFamily
	Attributes           map[string]string
	ReadReplicaSettings  ReadReplicasSettings
	StorageSettings      StorageSettings
	KeyBloomFilter       FeatureFlag
	PartitioningSettings PartitioningSettings
	Indexes              []IndexDescription
	TimeToLiveSettings   *TimeToLiveSettings
	Changefeeds          []ChangefeedDescription
	Tiering              string
	StoreType            StoreType
}

type TableStats struct {
	PartitionStats   []PartitionStats
	RowsEstimate     uint64
	StoreSize        uint64
	Partitions       uint64
	CreationTime     time.Time
	ModificationTime time.Time
}

type PartitionStats struct {
	RowsEstimate uint64
	StoreSize    uint64
	LeaderNodeID uint32
}

type ColumnFamily struct {
	Name         string
	Data         StoragePool
	Compression  ColumnFamilyCompression
	KeepInMemory FeatureFlag
}

func (c ColumnFamily) toYDB() *Ydb_Table.ColumnFamily {
	return &Ydb_Table.ColumnFamily{
		Name:         c.Name,
		Data:         c.Data.toYDB(),
		Compression:  c.Compression.toYDB(),
		KeepInMemory: c.KeepInMemory.ToYDB(),
	}
}

func NewColumnFamily(c *Ydb_Table.ColumnFamily) ColumnFamily {
	return ColumnFamily{
		Name:         c.GetName(),
		Data:         storagePool(c.GetData()),
		Compression:  columnFamilyCompression(c.GetCompression()),
		KeepInMemory: feature.FromYDB(c.GetKeepInMemory()),
	}
}

type StoragePool struct {
	Media string
}

func (s StoragePool) toYDB() *Ydb_Table.StoragePool {
	if s.Media == "" {
		return nil
	}

	return &Ydb_Table.StoragePool{
		Media: s.Media,
	}
}

func storagePool(s *Ydb_Table.StoragePool) StoragePool {
	return StoragePool{
		Media: s.GetMedia(),
	}
}

type ColumnFamilyCompression byte

const (
	ColumnFamilyCompressionUnknown ColumnFamilyCompression = iota
	ColumnFamilyCompressionNone
	ColumnFamilyCompressionLZ4
)

func (c ColumnFamilyCompression) String() string {
	switch c {
	case ColumnFamilyCompressionNone:
		return "none"
	case ColumnFamilyCompressionLZ4:
		return "lz4"
	default:
		return fmt.Sprintf("unknown_column_family_compression_%d", c)
	}
}

func (c ColumnFamilyCompression) toYDB() Ydb_Table.ColumnFamily_Compression {
	switch c {
	case ColumnFamilyCompressionNone:
		return Ydb_Table.ColumnFamily_COMPRESSION_NONE
	case ColumnFamilyCompressionLZ4:
		return Ydb_Table.ColumnFamily_COMPRESSION_LZ4
	default:
		return Ydb_Table.ColumnFamily_COMPRESSION_UNSPECIFIED
	}
}

func columnFamilyCompression(c Ydb_Table.ColumnFamily_Compression) ColumnFamilyCompression {
	switch c {
	case Ydb_Table.ColumnFamily_COMPRESSION_NONE:
		return ColumnFamilyCompressionNone
	case Ydb_Table.ColumnFamily_COMPRESSION_LZ4:
		return ColumnFamilyCompressionLZ4
	default:
		return ColumnFamilyCompressionUnknown
	}
}

type (
	DescribeTableDesc   Ydb_Table.DescribeTableRequest
	DescribeTableOption func(d *DescribeTableDesc)
)

type ReadReplicasSettings struct {
	Type  ReadReplicasType
	Count uint64
}

func (rr ReadReplicasSettings) ToYDB() *Ydb_Table.ReadReplicasSettings {
	switch rr.Type {
	case ReadReplicasPerAzReadReplicas:
		return &Ydb_Table.ReadReplicasSettings{
			Settings: &Ydb_Table.ReadReplicasSettings_PerAzReadReplicasCount{
				PerAzReadReplicasCount: rr.Count,
			},
		}

	default:
		return &Ydb_Table.ReadReplicasSettings{
			Settings: &Ydb_Table.ReadReplicasSettings_AnyAzReadReplicasCount{
				AnyAzReadReplicasCount: rr.Count,
			},
		}
	}
}

func NewReadReplicasSettings(rr *Ydb_Table.ReadReplicasSettings) ReadReplicasSettings {
	t := ReadReplicasPerAzReadReplicas
	var c uint64

	if c = rr.GetPerAzReadReplicasCount(); c != 0 {
		t = ReadReplicasPerAzReadReplicas
	} else if c = rr.GetAnyAzReadReplicasCount(); c != 0 {
		t = ReadReplicasAnyAzReadReplicas
	}

	return ReadReplicasSettings{
		Type:  t,
		Count: c,
	}
}

type ReadReplicasType byte

const (
	ReadReplicasPerAzReadReplicas ReadReplicasType = iota
	ReadReplicasAnyAzReadReplicas
)

type StorageSettings struct {
	TableCommitLog0    StoragePool
	TableCommitLog1    StoragePool
	External           StoragePool
	StoreExternalBlobs FeatureFlag
}

func (ss StorageSettings) ToYDB() *Ydb_Table.StorageSettings {
	return &Ydb_Table.StorageSettings{
		TabletCommitLog0:   ss.TableCommitLog0.toYDB(),
		TabletCommitLog1:   ss.TableCommitLog1.toYDB(),
		External:           ss.External.toYDB(),
		StoreExternalBlobs: ss.StoreExternalBlobs.ToYDB(),
	}
}

func NewStorageSettings(ss *Ydb_Table.StorageSettings) StorageSettings {
	return StorageSettings{
		TableCommitLog0:    storagePool(ss.GetTabletCommitLog0()),
		TableCommitLog1:    storagePool(ss.GetTabletCommitLog1()),
		External:           storagePool(ss.GetExternal()),
		StoreExternalBlobs: feature.FromYDB(ss.GetStoreExternalBlobs()),
	}
}

type Partitions interface {
	CreateTableOption

	isPartitions()
}

type PartitioningSettings struct {
	PartitionBy        []string
	PartitioningBySize FeatureFlag
	PartitionSizeMb    uint64
	PartitioningByLoad FeatureFlag
	MinPartitionsCount uint64
	MaxPartitionsCount uint64
}

func (ps PartitioningSettings) toYDB() *Ydb_Table.PartitioningSettings {
	return &Ydb_Table.PartitioningSettings{
		PartitioningBySize: ps.PartitioningBySize.ToYDB(),
		PartitionSizeMb:    ps.PartitionSizeMb,
		PartitioningByLoad: ps.PartitioningByLoad.ToYDB(),
		MinPartitionsCount: ps.MinPartitionsCount,
		MaxPartitionsCount: ps.MaxPartitionsCount,
	}
}

func NewPartitioningSettings(ps *Ydb_Table.PartitioningSettings) PartitioningSettings {
	return PartitioningSettings{
		PartitionBy:        ps.GetPartitionBy(),
		PartitioningBySize: feature.FromYDB(ps.GetPartitioningBySize()),
		PartitionSizeMb:    ps.GetPartitionSizeMb(),
		PartitioningByLoad: feature.FromYDB(ps.GetPartitioningByLoad()),
		MinPartitionsCount: ps.GetMinPartitionsCount(),
		MaxPartitionsCount: ps.GetMaxPartitionsCount(),
	}
}

type (
	IndexType uint8
)

const (
	IndexTypeGlobal = IndexType(iota)
	IndexTypeGlobalAsync
)

func (t IndexType) ApplyIndexOption(d *indexDesc) {
	switch t {
	case IndexTypeGlobal:
		d.Type = &Ydb_Table.TableIndex_GlobalIndex{
			GlobalIndex: &Ydb_Table.GlobalIndex{},
		}
	case IndexTypeGlobalAsync:
		d.Type = &Ydb_Table.TableIndex_GlobalAsyncIndex{
			GlobalAsyncIndex: &Ydb_Table.GlobalAsyncIndex{},
		}
	}
}

func GlobalIndex() IndexType {
	return IndexTypeGlobal
}

func GlobalAsyncIndex() IndexType {
	return IndexTypeGlobalAsync
}

type PartitioningMode byte

const (
	PartitioningUnknown PartitioningMode = iota
	PartitioningDisabled
	PartitioningAutoSplit
	PartitioningAutoSplitMerge
)

func (p PartitioningMode) toYDB() Ydb_Table.PartitioningPolicy_AutoPartitioningPolicy {
	switch p {
	case PartitioningDisabled:
		return Ydb_Table.PartitioningPolicy_DISABLED
	case PartitioningAutoSplit:
		return Ydb_Table.PartitioningPolicy_AUTO_SPLIT
	case PartitioningAutoSplitMerge:
		return Ydb_Table.PartitioningPolicy_AUTO_SPLIT_MERGE
	default:
		panic("ydb: unknown partitioning mode")
	}
}

type ExecuteScanQueryRequestMode byte

const (
	ExecuteScanQueryRequestModeExec ExecuteScanQueryRequestMode = iota
	ExecuteScanQueryRequestModeExplain
)

func (p ExecuteScanQueryRequestMode) toYDB() Ydb_Table.ExecuteScanQueryRequest_Mode {
	switch p {
	case ExecuteScanQueryRequestModeExec:
		return Ydb_Table.ExecuteScanQueryRequest_MODE_EXEC
	case ExecuteScanQueryRequestModeExplain:
		return Ydb_Table.ExecuteScanQueryRequest_MODE_EXPLAIN
	default:
		panic("ydb: unknown execute scan query mode")
	}
}

type TableOptionsDescription struct {
	TableProfilePresets       []TableProfileDescription
	StoragePolicyPresets      []StoragePolicyDescription
	CompactionPolicyPresets   []CompactionPolicyDescription
	PartitioningPolicyPresets []PartitioningPolicyDescription
	ExecutionPolicyPresets    []ExecutionPolicyDescription
	ReplicationPolicyPresets  []ReplicationPolicyDescription
	CachingPolicyPresets      []CachingPolicyDescription
}

type (
	TableProfileDescription struct {
		Name   string
		Labels map[string]string

		DefaultStoragePolicy      string
		DefaultCompactionPolicy   string
		DefaultPartitioningPolicy string
		DefaultExecutionPolicy    string
		DefaultReplicationPolicy  string
		DefaultCachingPolicy      string

		AllowedStoragePolicies      []string
		AllowedCompactionPolicies   []string
		AllowedPartitioningPolicies []string
		AllowedExecutionPolicies    []string
		AllowedReplicationPolicies  []string
		AllowedCachingPolicies      []string
	}
	StoragePolicyDescription struct {
		Name   string
		Labels map[string]string
	}
	CompactionPolicyDescription struct {
		Name   string
		Labels map[string]string
	}
	PartitioningPolicyDescription struct {
		Name   string
		Labels map[string]string
	}
	ExecutionPolicyDescription struct {
		Name   string
		Labels map[string]string
	}
	ReplicationPolicyDescription struct {
		Name   string
		Labels map[string]string
	}
	CachingPolicyDescription struct {
		Name   string
		Labels map[string]string
	}
)

type KeyRange struct {
	From value.Value
	To   value.Value
}

func (kr KeyRange) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	if kr.From == nil {
		buf.WriteString("NULL")
	} else {
		buf.WriteString(kr.From.Yql())
	}
	buf.WriteString(",")
	if kr.To == nil {
		buf.WriteString("NULL")
	} else {
		buf.WriteString(kr.To.Yql())
	}
	buf.WriteString("]")

	return buf.String()
}

// Deprecated: use TimeToLiveSettings instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
type TTLSettings struct {
	DateTimeColumn string
	TTLSeconds     uint32
}

type TimeToLiveSettings struct {
	ColumnName string

	// Mode specified mode
	Mode TimeToLiveMode

	// ExpireAfterSeconds specified expiration in seconds
	ExpireAfterSeconds uint32

	// ColumnUnit valid with Mode = TimeToLiveModeValueSinceUnixEpoch
	ColumnUnit *TimeToLiveUnit
}

type TimeToLiveMode byte

const (
	TimeToLiveModeDateType TimeToLiveMode = iota
	TimeToLiveModeValueSinceUnixEpoch
)

func NewTTLSettings() TimeToLiveSettings {
	return TimeToLiveSettings{
		Mode: TimeToLiveModeDateType,
	}
}

func (ttl TimeToLiveSettings) ColumnDateType(columnName string) TimeToLiveSettings {
	ttl.Mode = TimeToLiveModeDateType
	ttl.ColumnName = columnName

	return ttl
}

func unitToPointer(unit TimeToLiveUnit) *TimeToLiveUnit {
	return &unit
}

func (ttl TimeToLiveSettings) ColumnSeconds(columnName string) TimeToLiveSettings {
	ttl.Mode = TimeToLiveModeValueSinceUnixEpoch
	ttl.ColumnName = columnName
	ttl.ColumnUnit = unitToPointer(TimeToLiveUnitSeconds)

	return ttl
}

func (ttl TimeToLiveSettings) ColumnMilliseconds(columnName string) TimeToLiveSettings {
	ttl.Mode = TimeToLiveModeValueSinceUnixEpoch
	ttl.ColumnName = columnName
	ttl.ColumnUnit = unitToPointer(TimeToLiveUnitMilliseconds)

	return ttl
}

func (ttl TimeToLiveSettings) ColumnMicroseconds(columnName string) TimeToLiveSettings {
	ttl.Mode = TimeToLiveModeValueSinceUnixEpoch
	ttl.ColumnName = columnName
	ttl.ColumnUnit = unitToPointer(TimeToLiveUnitMicroseconds)

	return ttl
}

func (ttl TimeToLiveSettings) ColumnNanoseconds(columnName string) TimeToLiveSettings {
	ttl.Mode = TimeToLiveModeValueSinceUnixEpoch
	ttl.ColumnName = columnName
	ttl.ColumnUnit = unitToPointer(TimeToLiveUnitNanoseconds)

	return ttl
}

func (ttl TimeToLiveSettings) ExpireAfter(expireAfter time.Duration) TimeToLiveSettings {
	ttl.ExpireAfterSeconds = uint32(expireAfter.Seconds())

	return ttl
}

func (ttl *TimeToLiveSettings) ToYDB() *Ydb_Table.TtlSettings {
	if ttl == nil {
		return nil
	}
	switch ttl.Mode {
	case TimeToLiveModeValueSinceUnixEpoch:
		return &Ydb_Table.TtlSettings{
			Mode: &Ydb_Table.TtlSettings_ValueSinceUnixEpoch{
				ValueSinceUnixEpoch: &Ydb_Table.ValueSinceUnixEpochModeSettings{
					ColumnName:         ttl.ColumnName,
					ColumnUnit:         ttl.ColumnUnit.ToYDB(),
					ExpireAfterSeconds: ttl.ExpireAfterSeconds,
				},
			},
		}
	default: // currently use TimeToLiveModeDateType mode as default
		return &Ydb_Table.TtlSettings{
			Mode: &Ydb_Table.TtlSettings_DateTypeColumn{
				DateTypeColumn: &Ydb_Table.DateTypeColumnModeSettings{
					ColumnName:         ttl.ColumnName,
					ExpireAfterSeconds: ttl.ExpireAfterSeconds,
				},
			},
		}
	}
}

type TimeToLiveUnit int32

const (
	TimeToLiveUnitUnspecified TimeToLiveUnit = iota
	TimeToLiveUnitSeconds
	TimeToLiveUnitMilliseconds
	TimeToLiveUnitMicroseconds
	TimeToLiveUnitNanoseconds
)

func (unit *TimeToLiveUnit) ToYDB() Ydb_Table.ValueSinceUnixEpochModeSettings_Unit {
	if unit == nil {
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_UNSPECIFIED
	}
	switch *unit {
	case TimeToLiveUnitSeconds:
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_SECONDS
	case TimeToLiveUnitMilliseconds:
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_MILLISECONDS
	case TimeToLiveUnitMicroseconds:
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_MICROSECONDS
	case TimeToLiveUnitNanoseconds:
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_NANOSECONDS
	case TimeToLiveUnitUnspecified:
		return Ydb_Table.ValueSinceUnixEpochModeSettings_UNIT_UNSPECIFIED
	default:
		panic("ydb: unknown unit for value since epoch")
	}
}

type ChangefeedDescription struct {
	Name             string
	Mode             ChangefeedMode
	Format           ChangefeedFormat
	State            ChangefeedState
	VirtualTimestamp bool
}

func NewChangefeedDescription(proto *Ydb_Table.ChangefeedDescription) ChangefeedDescription {
	return ChangefeedDescription{
		Name:             proto.GetName(),
		Mode:             ChangefeedMode(proto.GetMode()),
		Format:           ChangefeedFormat(proto.GetFormat()),
		State:            ChangefeedState(proto.GetState()),
		VirtualTimestamp: proto.GetVirtualTimestamps(),
	}
}

type ChangefeedState int

const (
	ChangefeedStateUnspecified = ChangefeedState(Ydb_Table.ChangefeedDescription_STATE_UNSPECIFIED)
	ChangefeedStateEnabled     = ChangefeedState(Ydb_Table.ChangefeedDescription_STATE_ENABLED)
	ChangefeedStateDisabled    = ChangefeedState(Ydb_Table.ChangefeedDescription_STATE_DISABLED)
)

type ChangefeedMode int

const (
	ChangefeedModeUnspecified     = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_UNSPECIFIED)
	ChangefeedModeKeysOnly        = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_KEYS_ONLY)
	ChangefeedModeUpdates         = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_UPDATES)
	ChangefeedModeNewImage        = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_NEW_IMAGE)
	ChangefeedModeOldImage        = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_OLD_IMAGE)
	ChangefeedModeNewAndOldImages = ChangefeedMode(Ydb_Table.ChangefeedMode_MODE_NEW_AND_OLD_IMAGES)
)

type ChangefeedFormat int

const (
	ChangefeedFormatUnspecified         = ChangefeedFormat(Ydb_Table.ChangefeedFormat_FORMAT_UNSPECIFIED)
	ChangefeedFormatJSON                = ChangefeedFormat(Ydb_Table.ChangefeedFormat_FORMAT_JSON)
	ChangefeedFormatDynamoDBStreamsJSON = ChangefeedFormat(Ydb_Table.ChangefeedFormat_FORMAT_DYNAMODB_STREAMS_JSON)
)

type StoreType int

const (
	StoreTypeUnspecified = StoreType(Ydb_Table.StoreType_STORE_TYPE_UNSPECIFIED)
	StoreTypeRow         = StoreType(Ydb_Table.StoreType_STORE_TYPE_ROW)
	StoreTypeColumn      = StoreType(Ydb_Table.StoreType_STORE_TYPE_COLUMN)
)
