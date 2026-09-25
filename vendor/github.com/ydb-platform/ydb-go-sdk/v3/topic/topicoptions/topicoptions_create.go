package topicoptions

import (
	"sort"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
)

// CreateOption type for options of topic create
type CreateOption interface {
	ApplyCreateOption(request *rawtopic.CreateTopicRequest)
}

// CreateWithMeteringMode set metering mode for the topic
func CreateWithMeteringMode(mode topictypes.MeteringMode) CreateOption {
	return withMeteringMode(mode)
}

// CreateWithMinActivePartitions set min active partitions for the topic
func CreateWithMinActivePartitions(count int64) CreateOption {
	return withMinActivePartitions(count)
}

// CreateWithPartitionCountLimit set partition count limit for the topic
// Deprecated: Use CreateWithMaxActivePartitions instead.
// Will be removed after Nov 2025.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func CreateWithPartitionCountLimit(count int64) CreateOption {
	return withPartitionCountLimit(count)
}

// CreateWithRetentionPeriod set retention time interval for the topic.
func CreateWithRetentionPeriod(retentionPeriod time.Duration) CreateOption {
	return withRetentionPeriod(retentionPeriod)
}

// CreateWithRetentionStorageMB set retention size for the topic.
func CreateWithRetentionStorageMB(retentionStorageMB int64) CreateOption {
	return withRetentionStorageMB(retentionStorageMB)
}

// CreateWithSupportedCodecs set supported codecs for the topic
func CreateWithSupportedCodecs(codecs ...topictypes.Codec) CreateOption {
	sort.Slice(codecs, func(i, j int) bool {
		return codecs[i] < codecs[j]
	})

	return withSupportedCodecs(codecs)
}

// CreateWithPartitionWriteSpeedBytesPerSecond set write size limit for partitions of the topic
func CreateWithPartitionWriteSpeedBytesPerSecond(partitionWriteSpeedBytesPerSecond int64) CreateOption {
	return withPartitionWriteSpeedBytesPerSecond(partitionWriteSpeedBytesPerSecond)
}

// CreateWithPartitionWriteBurstBytes set burst limit for partitions of the topic
func CreateWithPartitionWriteBurstBytes(partitionWriteBurstBytes int64) CreateOption {
	return withPartitionWriteBurstBytes(partitionWriteBurstBytes)
}

// CreateWithAttributes set attributes for the topic.
func CreateWithAttributes(attributes map[string]string) CreateOption {
	return withAttributes(attributes)
}

// CreateWithConsumer create new consumers with the topic
func CreateWithConsumer(consumers ...topictypes.Consumer) CreateOption {
	sort.Slice(consumers, func(i, j int) bool {
		return consumers[i].Name < consumers[j].Name
	})

	return withAddConsumers(consumers)
}

// CreateWithMaxActivePartitions set max active partitions for the topic
func CreateWithMaxActivePartitions(count int64) CreateOption {
	return withMaxActivePartitions(count)
}

// CreateWithAutoPartitioningSettings set auto partitioning settings for the topic
func CreateWithAutoPartitioningSettings(settings topictypes.AutoPartitioningSettings) CreateOption {
	return withAutoPartitioningSettings(settings)
}

type withMaxActivePartitions int64

func (count withMaxActivePartitions) ApplyCreateOption(request *rawtopic.CreateTopicRequest) {
	request.PartitioningSettings.MaxActivePartitions = int64(count)
}

type withAutoPartitioningSettings topictypes.AutoPartitioningSettings

func (settings withAutoPartitioningSettings) ApplyCreateOption(request *rawtopic.CreateTopicRequest) {
	s := topictypes.AutoPartitioningSettings(settings)
	s.ToRaw(&request.PartitioningSettings.AutoPartitioningSettings)
}
