package topic

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers/topic"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/attributes"
)

func flattenYDBTopicDescription(d *schema.ResourceData, desc topictypes.TopicDescription) error {
	_ = d.Set(attributeName, d.Get(attributeName).(string)) // NOTE(shmel1k@): TopicService SDK does not return path for stream.
	_ = d.Set(attributePartitionsCount, desc.PartitionSettings.MinActivePartitions)
	_ = d.Set(attributeMaxPartitionsCount, desc.PartitionSettings.MaxActivePartitions)
	_ = d.Set(attributeRetentionPeriodHours, desc.RetentionPeriod.Hours())
	_ = d.Set(attributeRetentionStorageMB, desc.RetentionStorageMB)
	_ = d.Set(attributeMeteringMode, MeteringModeToString(desc.MeteringMode))
	_ = d.Set(attributePartitionWriteSpeedKBPS, desc.PartitionWriteSpeedBytesPerSecond/1024)
	_ = d.Set(attributeAutoPartitioningSettings, []map[string]interface{}{
		{
			attributeAutoPartitioningStrategy: convertFromAutoPartitioningStrategy(desc.PartitionSettings.AutoPartitioningSettings.AutoPartitioningStrategy),
			attributeAutoPartitioningWriteSpeedStrategy: []map[string]interface{}{
				{
					attributeStabilizationWindow:    desc.PartitionSettings.AutoPartitioningSettings.AutoPartitioningWriteSpeedStrategy.StabilizationWindow / time.Second,
					attributeUpUtilizationPercent:   desc.PartitionSettings.AutoPartitioningSettings.AutoPartitioningWriteSpeedStrategy.UpUtilizationPercent,
					attributeDownUtilizationPercent: desc.PartitionSettings.AutoPartitioningSettings.AutoPartitioningWriteSpeedStrategy.DownUtilizationPercent,
				},
			},
		},
	})

	supportedCodecs := make([]string, 0, len(desc.SupportedCodecs))
	for _, v := range desc.SupportedCodecs {
		switch v {
		case topictypes.CodecRaw:
			supportedCodecs = append(supportedCodecs, ydbTopicCodecRAW)
		case topictypes.CodecZstd:
			supportedCodecs = append(supportedCodecs, ydbTopicCodecZSTD)
		case topictypes.CodecGzip:
			supportedCodecs = append(supportedCodecs, ydbTopicCodecGZIP)
		}
	}

	consumers := topic.FlattenConsumersDescription(desc.Consumers)
	err := d.Set(attributeConsumer, consumers)
	if err != nil {
		return fmt.Errorf("failed to set consumer %+v: %w", consumers, err)
	}

	err = d.Set(attributeSupportedCodecs, supportedCodecs)
	if err != nil {
		return err
	}

	return d.Set(attributes.DatabaseEndpoint, d.Get(attributes.DatabaseEndpoint).(string))
}

func prepareYDBTopicAlterSettings(
	d *schema.ResourceData,
	settings topictypes.TopicDescription,
) (opts []topicoptions.AlterOption) {
	if d.HasChange(attributePartitionsCount) {
		opts = append(opts, topicoptions.AlterWithPartitionCountLimit(int64(d.Get("partitions_count").(int))))
		opts = append(opts, topicoptions.AlterWithMinActivePartitions(int64(d.Get("partitions_count").(int))))
	}
	if d.HasChange(attributeMaxPartitionsCount) {
		opts = append(opts, topicoptions.AlterWithMaxActivePartitions(int64(d.Get(attributeMaxPartitionsCount).(int))))
	}
	if d.HasChange(attributeMeteringMode) {
		opts = append(opts, topicoptions.AlterWithMeteringMode(StringToMeteringMode(d.Get("metering_mode").(string))))
	}
	if d.HasChange(attributeSupportedCodecs) {
		codecs := d.Get(attributeSupportedCodecs).(*schema.Set)
		updatedCodecs := make([]topictypes.Codec, 0, len(codecs.List()))

		for _, c := range codecs.List() {
			cc, ok := topic.YDBTopicCodecNameToCodec[strings.ToLower(c.(string))]
			if !ok {
				panic(fmt.Sprintf("Unsupported codec %q found after validation", cc))
			}
			updatedCodecs = append(updatedCodecs, cc)
		}
		opts = append(opts, topicoptions.AlterWithSupportedCodecs(updatedCodecs...))
	}
	if d.HasChange(attributeRetentionPeriodHours) {
		opts = append(opts, topicoptions.AlterWithRetentionPeriod(time.Duration(d.Get(attributeRetentionPeriodHours).(int))*time.Hour))
	}
	if d.HasChange(attributeRetentionStorageMB) {
		opts = append(opts, topicoptions.AlterWithRetentionStorageMB(int64(d.Get(attributeRetentionStorageMB).(int))))
	}
	if d.HasChange(attributePartitionWriteSpeedKBPS) {
		writeSpeed := d.Get(attributePartitionWriteSpeedKBPS).(int) * 1024
		opts = append(opts, topicoptions.AlterWithPartitionWriteSpeedBytesPerSecond(int64(writeSpeed)))
	}
	if d.HasChange(attributeConsumer) {
		additionalOpts := topic.MergeConsumerSettings(d.Get(attributeConsumer).(*schema.Set), settings.Consumers)
		opts = append(opts, additionalOpts...)
	}
	if d.HasChange(attributeAutoPartitioningSettings) {
		autoPartitioningSettings := d.Get(attributeAutoPartitioningSettings).([]interface{})
		if len(autoPartitioningSettings) > 0 {
			settings := autoPartitioningSettings[0].(map[string]interface{})
			opts = append(opts, topicoptions.AlterWithAutoPartitioningStrategy(convertToAutoPartitioningStrategy(settings[attributeAutoPartitioningStrategy].(string))))
			speedStrategy := settings[attributeAutoPartitioningWriteSpeedStrategy].([]interface{})
			if len(speedStrategy) > 0 {
				writeSpeedStrategy := speedStrategy[0].(map[string]interface{})
				opts = append(opts, topicoptions.AlterWithAutoPartitioningWriteSpeedStabilizationWindow(time.Duration(writeSpeedStrategy[attributeStabilizationWindow].(int))*time.Second))
				opts = append(opts, topicoptions.AlterWithAutoPartitioningWriteSpeedUpUtilizationPercent(int32(writeSpeedStrategy[attributeUpUtilizationPercent].(int))))
				opts = append(opts, topicoptions.AlterWithAutoPartitioningWriteSpeedDownUtilizationPercent(int32(writeSpeedStrategy[attributeDownUtilizationPercent].(int))))
			}
		}
	}

	return opts
}
