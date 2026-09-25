package changefeed

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers/topic"
	tbl "github.com/ydb-platform/terraform-provider-ydb/internal/table"
)

func (h *handler) Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cdcResource, err := changefeedResourceSchemaToChangefeedResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := helpers.AreAllElementsUnique(cdcResource.Consumers); err != nil {
		return diag.FromErr(err)
	}

	db, err := tbl.CreateDBConnection(ctx, tbl.ClientParams{
		DatabaseEndpoint: cdcResource.getConnectionString(),
		AuthCreds:        h.authCreds,
	})
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "failed to initialize table client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = db.Close(ctx)
	}()

	topicPath := helpers.TrimPath(cdcResource.getTablePath()) + "/" + cdcResource.Name
	desc, err := db.Topic().Describe(ctx, topicPath)
	if err != nil {
		return diag.FromErr(err)
	}

	alterConsumersOptions := mergeConsumerSettings(d, desc.Consumers)
	err = db.Topic().Alter(ctx, topicPath, alterConsumersOptions...)
	if err != nil {
		return diag.FromErr(err)
	}

	return h.Read(ctx, d, meta)
}

func mergeConsumerSettings(d *schema.ResourceData, readRules []topictypes.Consumer) (opts []topicoptions.AlterOption) {
	rules := make(map[string]topictypes.Consumer, len(readRules))
	for i := 0; i < len(readRules); i++ {
		rules[readRules[i].Name] = readRules[i]
	}

	// TODO(shmel1k@): remove copypaste
	consumersMap := make(map[string]struct{})

	pList := d.Get("consumer").([]interface{})
	for _, v := range pList {
		consumer := v.(map[string]interface{})
		consumerName, ok := consumer["name"].(string)
		if !ok {
			continue
		}

		consumersMap[consumerName] = struct{}{}

		supportedCodecs, ok := consumer["supported_codecs"].([]interface{})
		if !ok {
			for _, vv := range topic.YDBTopicAllowedCodecs {
				supportedCodecs = append(supportedCodecs, vv)
			}
		}
		startingMessageTS, ok := consumer["starting_message_timestamp_ms"].(int)
		if !ok {
			startingMessageTS = 0
		}

		r, ok := rules[consumerName]
		if !ok {
			// consumer was deleted by someone outside terraform or does not exist.
			codecs := make([]topictypes.Codec, 0, len(supportedCodecs))
			for _, c := range supportedCodecs {
				codec := c.(string)
				codecs = append(codecs, topic.YDBTopicCodecNameToCodec[strings.ToLower(codec)])
			}
			opts = append(opts, topicoptions.AlterWithAddConsumers(
				topictypes.Consumer{
					Name:            consumerName,
					ReadFrom:        time.UnixMilli(int64(startingMessageTS)),
					SupportedCodecs: codecs,
				},
			))
			continue
		}

		readFrom := time.UnixMilli(int64(startingMessageTS))
		if r.ReadFrom != readFrom {
			opts = append(opts, topicoptions.AlterConsumerWithReadFrom(consumerName, readFrom))
		}

		newCodecs := make([]topictypes.Codec, 0, len(supportedCodecs))
		for _, codec := range supportedCodecs {
			c := topic.YDBTopicCodecNameToCodec[strings.ToLower(codec.(string))]
			newCodecs = append(newCodecs, c)
		}
		if len(newCodecs) != 0 {
			opts = append(opts, topicoptions.AlterConsumerWithSupportedCodecs(consumerName, newCodecs))
		}
	}
	for k := range rules {
		if _, ok := consumersMap[k]; !ok {
			opts = append(opts, topicoptions.AlterWithDropConsumers(k))
		}
	}
	return opts
}
