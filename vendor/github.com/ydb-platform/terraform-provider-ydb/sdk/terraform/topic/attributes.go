package topic

const (
	attributePartitionsCount                        = "partitions_count"
	attributeMaxPartitionsCount                     = "max_partitions_count"
	attributeMeteringMode                           = "metering_mode"
	attributeSupportedCodecs                        = "supported_codecs"
	attributeRetentionPeriodHours                   = "retention_period_hours"
	attributeRetentionStorageMB                     = "retention_storage_mb"
	attributePartitionWriteSpeedKBPS                = "partition_write_speed_kbps"
	attributeConsumer                               = "consumer"
	attributeName                                   = "name" // NOTE(shmel1k@): deprecated, use 'attributes.Path' instead.
	attributeConsumerStartingMessageTimestampMS     = "starting_message_timestamp_ms"
	attributeConsumerImportant                      = "important"
	attributeAutoPartitioningSettings               = "auto_partitioning_settings"
	attributeAutoPartitioningStrategy               = "auto_partitioning_strategy"
	attributeAutoPartitioningStrategyUnspecified    = "UNSPECIFIED"
	attributeAutoPartitioningStrategyDisabled       = "DISABLED"
	attributeAutoPartitioningStrategyScaleUp        = "SCALE_UP"
	attributeAutoPartitioningStrategyScaleUpAndDown = "SCALE_UP_AND_DOWN"
	attributeAutoPartitioningStrategyPaused         = "PAUSED"
	attributeAutoPartitioningWriteSpeedStrategy     = "auto_partitioning_write_speed_strategy"
	attributeStabilizationWindow                    = "stabilization_window"
	attributeUpUtilizationPercent                   = "up_utilization_percent"
	attributeDownUtilizationPercent                 = "down_utilization_percent"
)
