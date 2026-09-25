package topicoptions

import (
	"sort"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
)

// AlterOption type of options for change topic settings
type AlterOption interface {
	ApplyAlterOption(req *rawtopic.AlterTopicRequest)
}

// AlterWithMeteringMode change metering mode for topic (need for serverless installations)
func AlterWithMeteringMode(m topictypes.MeteringMode) AlterOption {
	return withMeteringMode(m)
}

// AlterWithMinActivePartitions change min active partitions of the topic
func AlterWithMinActivePartitions(minActivePartitions int64) AlterOption {
	return withMinActivePartitions(minActivePartitions)
}

// AlterWithPartitionCountLimit change partition count limit of the topic
func AlterWithPartitionCountLimit(partitionCountLimit int64) AlterOption {
	return withPartitionCountLimit(partitionCountLimit)
}

// AlterWithRetentionPeriod change retention period of topic
func AlterWithRetentionPeriod(retentionPeriod time.Duration) AlterOption {
	return withRetentionPeriod(retentionPeriod)
}

// AlterWithRetentionStorageMB change retention storage size in MB.
func AlterWithRetentionStorageMB(retentionStorageMB int64) AlterOption {
	return withRetentionStorageMB(retentionStorageMB)
}

// AlterWithSupportedCodecs change set of codec, allowed for the topic
func AlterWithSupportedCodecs(codecs ...topictypes.Codec) AlterOption {
	sort.Slice(codecs, func(i, j int) bool {
		return codecs[i] < codecs[j]
	})

	return withSupportedCodecs(codecs)
}

// AlterWithPartitionWriteSpeedBytesPerSecond change limit of write speed for partitions of the topic
func AlterWithPartitionWriteSpeedBytesPerSecond(bytesPerSecond int64) AlterOption {
	return withPartitionWriteSpeedBytesPerSecond(bytesPerSecond)
}

// AlterWithPartitionWriteBurstBytes change burst size for write to partition of topic
func AlterWithPartitionWriteBurstBytes(burstBytes int64) AlterOption {
	return withPartitionWriteBurstBytes(burstBytes)
}

// AlterWithAttributes change attributes map of topic
func AlterWithAttributes(attributes map[string]string) AlterOption {
	return withAttributes(attributes)
}

// AlterWithAddConsumers add consumer to the topic
func AlterWithAddConsumers(consumers ...topictypes.Consumer) AlterOption {
	sort.Slice(consumers, func(i, j int) bool {
		return consumers[i].Name < consumers[j].Name
	})

	return withAddConsumers(consumers)
}

// AlterWithDropConsumers drop consumer from the topic
func AlterWithDropConsumers(consumersName ...string) AlterOption {
	sort.Strings(consumersName)

	return withDropConsumers(consumersName)
}

// AlterConsumerWithImportant set/remove important flag for the consumer of topic
func AlterConsumerWithImportant(name string, important bool) AlterOption {
	return withConsumerWithImportant{
		name:      name,
		important: important,
	}
}

// AlterConsumerWithReadFrom change min time of messages, received for the topic
func AlterConsumerWithReadFrom(name string, readFrom time.Time) AlterOption {
	return withConsumerWithReadFrom{
		name:     name,
		readFrom: readFrom,
	}
}

// AlterConsumerWithSupportedCodecs change codecs, supported by the consumer
func AlterConsumerWithSupportedCodecs(name string, codecs []topictypes.Codec) AlterOption {
	sort.Slice(codecs, func(i, j int) bool {
		return codecs[i] < codecs[j]
	})

	return withConsumerWithSupportedCodecs{
		name:   name,
		codecs: codecs,
	}
}

// AlterConsumerWithAttributes change attributes of the consumer
func AlterConsumerWithAttributes(name string, attributes map[string]string) AlterOption {
	return withConsumerWithAttributes{
		name:       name,
		attributes: attributes,
	}
}

// AlterWithMaxActivePartitions change max active partitions of the topic
func AlterWithMaxActivePartitions(maxActivePartitions int64) AlterOption {
	return alterWithMaxActivePartitions(maxActivePartitions)
}

// AlterWithAutoPartitioningStrategy change auto partitioning strategy for the topic
func AlterWithAutoPartitioningStrategy(strategy topictypes.AutoPartitioningStrategy) AlterOption {
	return withAutoPartitioningStrategy(strategy)
}

// AlterWithAutoPartitioningWriteSpeedStabilizationWindow change stabilization window for auto partitioning write speed
func AlterWithAutoPartitioningWriteSpeedStabilizationWindow(window time.Duration) AlterOption {
	return withAutoPartitioningWriteSpeedStabilizationWindow(window)
}

// AlterWithAutoPartitioningWriteSpeedUpUtilizationPercent change up utilization percent for auto
// partitioning write speed
func AlterWithAutoPartitioningWriteSpeedUpUtilizationPercent(percent int32) AlterOption {
	return withAutoPartitioningWriteSpeedUpUtilizationPercent(percent)
}

// AlterWithAutoPartitioningWriteSpeedDownUtilizationPercent change down utilization percent for auto
// partitioning write speed
func AlterWithAutoPartitioningWriteSpeedDownUtilizationPercent(percent int32) AlterOption {
	return withAutoPartitioningWriteSpeedDownUtilizationPercent(percent)
}

func ensureAlterConsumer(
	consumers []rawtopic.AlterConsumer,
	name string,
) (
	newConsumers []rawtopic.AlterConsumer,
	index int,
) {
	for i := range consumers {
		if consumers[i].Name == name {
			return consumers, i
		}
	}
	consumers = append(consumers, rawtopic.AlterConsumer{Name: name})

	return consumers, len(consumers) - 1
}

type alterWithMaxActivePartitions int64

func (maxActivePartitions alterWithMaxActivePartitions) ApplyAlterOption(req *rawtopic.AlterTopicRequest) {
	req.AlterPartitionSettings.SetMaxActivePartitions.HasValue = true
	req.AlterPartitionSettings.SetMaxActivePartitions.Value = int64(maxActivePartitions)
}

type withAutoPartitioningStrategy topictypes.AutoPartitioningStrategy

func (strategy withAutoPartitioningStrategy) ApplyAlterOption(req *rawtopic.AlterTopicRequest) {
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings = &rawtopic.AlterAutoPartitioningSettings{}
	}
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetStrategy = rawtopic.AutoPartitioningStrategy(strategy)
}

type withAutoPartitioningWriteSpeedStabilizationWindow time.Duration

//nolint:lll
func (window withAutoPartitioningWriteSpeedStabilizationWindow) ApplyAlterOption(req *rawtopic.AlterTopicRequest) {
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings = &rawtopic.AlterAutoPartitioningSettings{}
	}
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed = &rawtopic.AlterAutoPartitioningWriteSpeedStrategy{}
	}
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetStabilizationWindow.HasValue = true
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetStabilizationWindow.Value = time.Duration(window)
}

type withAutoPartitioningWriteSpeedUpUtilizationPercent int32

//nolint:lll
func (percent withAutoPartitioningWriteSpeedUpUtilizationPercent) ApplyAlterOption(req *rawtopic.AlterTopicRequest) {
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings = &rawtopic.AlterAutoPartitioningSettings{}
	}
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed = &rawtopic.AlterAutoPartitioningWriteSpeedStrategy{}
	}
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetUpUtilizationPercent.HasValue = true
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetUpUtilizationPercent.Value = int32(percent)
}

type withAutoPartitioningWriteSpeedDownUtilizationPercent int32

//nolint:lll
func (percent withAutoPartitioningWriteSpeedDownUtilizationPercent) ApplyAlterOption(req *rawtopic.AlterTopicRequest) {
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings = &rawtopic.AlterAutoPartitioningSettings{}
	}
	if req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed == nil {
		req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed = &rawtopic.AlterAutoPartitioningWriteSpeedStrategy{}
	}
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetDownUtilizationPercent.HasValue = true
	req.AlterPartitionSettings.AlterAutoPartitioningSettings.SetPartitionWriteSpeed.SetDownUtilizationPercent.Value = int32(percent)
}
