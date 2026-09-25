package topic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ydb "github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers/topic"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/attributes"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
)

type caller struct {
	authCreds auth.YdbCredentials
}

const (
	meteringModeRequestUnits     = "request_units"
	meteringModeReservedCapacity = "reserved_capacity"
	meteringModeUnspecified      = "unspecified"
)

func (c *caller) createYDBConnection(
	ctx context.Context,
	d helpers.ResourceDataProxy,
	ydbEn *helpers.YDBEntity,
) (*ydb.Driver, error) {
	// TODO(shmel1k@): move to other level.
	var opts []ydb.Option
	var databaseEndpoint string
	if ydbEn != nil {
		databaseEndpoint = ydbEn.PrepareFullYDBEndpoint()
	} else {
		// NOTE(shmel1k@): resource is not initialized yet.
		databaseEndpoint = d.Get(attributes.DatabaseEndpoint).(string)
	}

	switch {
	case c.authCreds.Token != "":
		opts = append(opts, ydb.WithAccessTokenCredentials(c.authCreds.Token))
	case c.authCreds.User != "":
		opts = append(opts, ydb.WithStaticCredentials(c.authCreds.User, c.authCreds.Password))
	}

	sess, err := ydb.Open(ctx, databaseEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create control-plane client: %w", err)
	}
	return sess, nil
}

func (c *caller) performYDBTopicUpdate(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	topic, err := helpers.ParseYDBEntityID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ydbClient, err := c.createYDBConnection(ctx, d, topic)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to initialize ydb-topic control plane client: %w", err))
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()

	topicClient := ydbClient.Topic()

	topicName := topic.GetEntityPath()
	desc, err := topicClient.Describe(ctx, topicName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return c.resourceYDBTopicCreate(ctx, d, nil)
		}
		return diag.FromErr(fmt.Errorf("failed to get description for topic %q", topicName))
	}

	opts := prepareYDBTopicAlterSettings(d, desc)
	err = topicClient.Alter(ctx, topicName, opts...)
	if err != nil {
		return diag.FromErr(fmt.Errorf("got error when tried to alter topic: %w", err))
	}

	return c.resourceYDBTopicRead(ctx, d, nil)
}

func MeteringModeToString(mode topictypes.MeteringMode) string {
	if mode == topictypes.MeteringModeRequestUnits {
		return meteringModeRequestUnits
	}
	if mode == topictypes.MeteringModeReservedCapacity {
		return meteringModeReservedCapacity
	}
	return meteringModeUnspecified
}

// Helper function to convert string to AutoPartitioningStrategy
func convertToAutoPartitioningStrategy(s string) topictypes.AutoPartitioningStrategy {
	switch s {
	case "UNSPECIFIED":
		return topictypes.AutoPartitioningStrategyUnspecified
	case "DISABLED":
		return topictypes.AutoPartitioningStrategyDisabled
	case "SCALE_UP":
		return topictypes.AutoPartitioningStrategyScaleUp
	case "SCALE_UP_AND_DOWN":
		return topictypes.AutoPartitioningStrategyScaleUpAndDown
	case "PAUSED":
		return topictypes.AutoPartitioningStrategyPaused
	default:
		return topictypes.AutoPartitioningStrategyUnspecified
	}
}

// Helper function to convert AutoPartitioningStrategy to string
func convertFromAutoPartitioningStrategy(strategy topictypes.AutoPartitioningStrategy) string {
	switch strategy {
	case topictypes.AutoPartitioningStrategyUnspecified:
		return "UNSPECIFIED"
	case topictypes.AutoPartitioningStrategyDisabled:
		return "DISABLED"
	case topictypes.AutoPartitioningStrategyScaleUp:
		return "SCALE_UP"
	case topictypes.AutoPartitioningStrategyScaleUpAndDown:
		return "SCALE_UP_AND_DOWN"
	case topictypes.AutoPartitioningStrategyPaused:
		return "PAUSED"
	default:
		return "UNSPECIFIED"
	}
}

func (c *caller) resourceYDBTopicRead(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	topic, err := helpers.ParseYDBEntityID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ydbClient, err := c.createYDBConnection(ctx, d, topic)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to initialize ydb-topic control plane client: %w", err))
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()
	topicClient := ydbClient.Topic()

	topicName := topic.GetEntityPath()

	description, err := topicClient.Describe(ctx, topicName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			d.SetId("") // marking as non-existing resource.
			return nil
		}
		return diag.FromErr(fmt.Errorf("resource: failed to describe topic: %w", err))
	}
	err = flattenYDBTopicDescription(d, description)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to flatten topic description: %w", err))
	}

	return nil
}

func StringToMeteringMode(mode string) topictypes.MeteringMode {
	if mode == meteringModeRequestUnits {
		return topictypes.MeteringModeRequestUnits
	}
	if mode == meteringModeReservedCapacity {
		return topictypes.MeteringModeReservedCapacity
	}
	return topictypes.MeteringModeUnspecified
}

func (c *caller) resourceYDBTopicCreate(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client, err := c.createYDBConnection(ctx, d, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to initialize yds control plane client: %w", err))
	}
	defer func() {
		_ = client.Close(ctx)
	}()

	var supportedCodecs []topictypes.Codec
	if gotCodecs, ok := d.GetOk(attributeSupportedCodecs); !ok {
		supportedCodecs = topic.YDBTopicDefaultCodecs
	} else {
		for _, c := range gotCodecs.(*schema.Set).List() {
			cod := c.(string)
			supportedCodecs = append(supportedCodecs, topic.YDBTopicCodecNameToCodec[cod])
		}
	}

	autoPartitioningTopicOptions := topictypes.AutoPartitioningSettings{}
	autoPartitioningSettings := d.Get(attributeAutoPartitioningSettings).([]interface{})
	if len(autoPartitioningSettings) > 0 {
		settings := autoPartitioningSettings[0].(map[string]interface{})
		autoPartitioningTopicOptions.AutoPartitioningStrategy = convertToAutoPartitioningStrategy(settings[attributeAutoPartitioningStrategy].(string))
		speedStrategy := settings[attributeAutoPartitioningWriteSpeedStrategy].([]interface{})
		if len(speedStrategy) > 0 {
			writeSpeedStrategy := speedStrategy[0].(map[string]interface{})
			autoPartitioningTopicOptions.AutoPartitioningWriteSpeedStrategy = topictypes.AutoPartitioningWriteSpeedStrategy{
				StabilizationWindow:    time.Duration(writeSpeedStrategy[attributeStabilizationWindow].(int)) * time.Second,
				UpUtilizationPercent:   int32(writeSpeedStrategy[attributeUpUtilizationPercent].(int)),
				DownUtilizationPercent: int32(writeSpeedStrategy[attributeDownUtilizationPercent].(int)),
			}
		}
	}

	consumers := topic.ExpandConsumers(d.Get(attributeConsumer).(*schema.Set))
	options := []topicoptions.CreateOption{
		topicoptions.CreateWithSupportedCodecs(supportedCodecs...),
		topicoptions.CreateWithMinActivePartitions(int64(d.Get(attributePartitionsCount).(int))),
		topicoptions.CreateWithMaxActivePartitions(int64(d.Get(attributeMaxPartitionsCount).(int))),
		topicoptions.CreateWithAutoPartitioningSettings(autoPartitioningTopicOptions),
		topicoptions.CreateWithConsumer(consumers...),
	}
	if d.Get(attributeRetentionPeriodHours) != 0 {
		options = append(options, topicoptions.CreateWithRetentionPeriod(time.Duration(d.Get(attributeRetentionPeriodHours).(int))*time.Hour))
	}
	if d.Get(attributeRetentionStorageMB) != 0 {
		options = append(options, topicoptions.CreateWithRetentionStorageMB(int64(d.Get(attributeRetentionStorageMB).(int))))
	}
	if d.Get(attributeMeteringMode) != "" {
		options = append(options, topicoptions.CreateWithMeteringMode(StringToMeteringMode(d.Get(attributeMeteringMode).(string))))
	}
	if d.Get(attributePartitionWriteSpeedKBPS) != 0 {
		writeSpeed := 1024 * d.Get(attributePartitionWriteSpeedKBPS).(int)
		options = append(options, topicoptions.CreateWithPartitionWriteBurstBytes(int64(writeSpeed)))
		options = append(options, topicoptions.CreateWithPartitionWriteSpeedBytesPerSecond(int64(writeSpeed)))
	}
	err = client.Topic().Create(ctx, d.Get(attributeName).(string), options...)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to initialize ydb-topic control plane client: %w", err))
	}

	topicPath := d.Get(attributeName).(string)
	d.SetId(d.Get(attributes.DatabaseEndpoint).(string) + "?path=" + topicPath)

	return c.resourceYDBTopicRead(ctx, d, nil)
}

func (c *caller) resourceYDBTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_ = meta
	return c.performYDBTopicUpdate(ctx, d)
}

func (c *caller) resourceYDBTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_ = meta
	topic, err := helpers.ParseYDBEntityID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client, err := c.createYDBConnection(ctx, d, topic)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to initialize ydb-topic control plane client: %w", err))
	}
	defer func() {
		_ = client.Close(ctx)
	}()

	topicName := topic.GetEntityPath()
	err = client.Topic().Drop(ctx, topicName)
	return diag.FromErr(err)
}
