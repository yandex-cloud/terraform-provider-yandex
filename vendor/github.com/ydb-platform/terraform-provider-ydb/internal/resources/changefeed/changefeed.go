package changefeed

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers/topic"
)

var (
	changefeedModeToStringMap = map[options.ChangefeedMode]string{
		options.ChangefeedModeUnspecified:     "",
		options.ChangefeedModeKeysOnly:        "KEYS_ONLY",
		options.ChangefeedModeNewImage:        "NEW_IMAGE",
		options.ChangefeedModeOldImage:        "OLD_IMAGE",
		options.ChangefeedModeNewAndOldImages: "NEW_AND_OLD_IMAGES",
		options.ChangefeedModeUpdates:         "UPDATES",
	}

	changefeedFormatToStringMap = map[options.ChangefeedFormat]string{
		options.ChangefeedFormatUnspecified: "",
		options.ChangefeedFormatJSON:        "JSON",
	}
)

type changeDataCaptureSettings struct {
	ConnectionString  string
	TablePath         string
	Name              string
	Mode              string
	Format            *string
	RetentionPeriod   *string
	VirtualTimestamps *bool
	Entity            *helpers.YDBEntity
	TableEntity       *helpers.YDBEntity
	Consumers         []topictypes.Consumer
}

func (c *changeDataCaptureSettings) getTablePath() string {
	if c.TablePath != "" {
		return c.TablePath
	}
	return c.TableEntity.GetEntityPath()
}

func (c *changeDataCaptureSettings) getConnectionString() string {
	if c.ConnectionString != "" {
		return c.ConnectionString
	}
	return c.TableEntity.PrepareFullYDBEndpoint()
}

func expandConsumers(d *schema.ResourceData) []topictypes.Consumer {
	v, ok := d.GetOk("consumer")
	if !ok {
		return nil
	}

	pList := v.([]interface{})
	result := make([]topictypes.Consumer, 0, len(pList))
	for _, l := range pList {
		consumer := l.(map[string]interface{})
		supportedCodecs, ok := consumer["supported_codecs"].([]interface{})
		if !ok {
			for _, vv := range topic.YDBTopicAllowedCodecs {
				supportedCodecs = append(supportedCodecs, vv)
			}
		}
		consumerName := consumer["name"].(string)
		startingMessageTS, ok := consumer["starting_message_timestamp_ms"].(int)
		if !ok {
			startingMessageTS = 0
		}
		codecs := make([]topictypes.Codec, 0, len(supportedCodecs))
		for _, c := range supportedCodecs {
			codec := c.(string)
			codecs = append(codecs, topic.YDBTopicCodecNameToCodec[strings.ToLower(codec)])
		}
		result = append(result, topictypes.Consumer{
			Name:            consumerName,
			SupportedCodecs: codecs,
			ReadFrom:        time.Unix(int64(startingMessageTS/1000), 0),
		})
	}

	return result
}

func changefeedResourceSchemaToChangefeedResource(d *schema.ResourceData) (*changeDataCaptureSettings, error) {
	var entity *helpers.YDBEntity
	var err error
	if d.Id() != "" {
		entity, err = helpers.ParseYDBEntityID(d.Id())
		if err != nil {
			return nil, fmt.Errorf("failed to parse changefeed entity: %w", err)
		}
	}

	var tableEntity *helpers.YDBEntity
	if tableID, ok := d.GetOk("table_id"); ok {
		en, err := helpers.ParseYDBEntityID(tableID.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to parse table_id: %w", err)
		}
		tableEntity = en
	}

	settings := &changeDataCaptureSettings{
		Entity:           entity,
		ConnectionString: d.Get("connection_string").(string),
		Name:             d.Get("name").(string),
		Mode:             d.Get("mode").(string),
		TablePath:        d.Get("table_path").(string),
		TableEntity:      tableEntity,
	}
	if format, ok := d.Get("format").(string); ok && format != "" {
		settings.Format = &format
	}
	if virtualTimestamps, ok := d.Get("virtual_timestamps").(bool); ok {
		settings.VirtualTimestamps = &virtualTimestamps
	}
	if retentionPeriod, ok := d.Get("retention_period").(string); ok && retentionPeriod != "" {
		settings.RetentionPeriod = &retentionPeriod
	}
	settings.Consumers = expandConsumers(d)

	return settings, nil
}

func flattenCDCDescription(
	d *schema.ResourceData,
	changefeedResource *changeDataCaptureSettings,
	cdcDescription options.ChangefeedDescription,
	consumers []topictypes.Consumer,
) (err error) {
	err = d.Set("table_path", changefeedResource.getTablePath())
	if err != nil {
		return
	}
	err = d.Set("connection_string", changefeedResource.getConnectionString())
	if err != nil {
		return
	}
	err = d.Set("table_id", changefeedResource.getConnectionString()+"?path="+changefeedResource.getTablePath())
	if err != nil {
		return
	}
	err = d.Set("name", cdcDescription.Name)
	if err != nil {
		return
	}
	err = d.Set("mode", changefeedModeToStringMap[cdcDescription.Mode])
	if err != nil {
		return
	}
	err = d.Set("format", changefeedFormatToStringMap[cdcDescription.Format])
	if err != nil {
		return
	}
	err = d.Set("virtual_timestamps", cdcDescription.VirtualTimestamp)
	if err != nil {
		return
	}

	curConsRaw := d.Get("consumer")
	cons := helpers.ConsumerSort(curConsRaw, consumers)

	// Записываем обновленных потребителей обратно в данные
	return d.Set("consumer", topic.FlattenConsumersDescription(cons))
}

func parseTablePathFromCDCEntity(entityPath string) string {
	split := strings.Split(entityPath, "/")
	return strings.Join(split[:len(split)-1], "/")
}
