package yandex

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_PersQueue_V1"

	"github.com/ydb-platform/ydb-go-persqueue-sdk/controlplane"
	"github.com/ydb-platform/ydb-go-persqueue-sdk/session"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ydbStreamCodecGZIP = "gzip"
	ydbStreamCodecRAW  = "raw"
	ydbStreamCodecZSTD = "zstd"
)

const (
	ydbStreamDefaultPartitionsCount = 2
	ydbStreamDefaultRetentionPeriod = 1000 * 60 * 60 * 24 // 1 day
)

var (
	ydbStreamAllowedCodecs = []string{
		ydbStreamCodecRAW,
		ydbStreamCodecGZIP,
		ydbStreamCodecZSTD,
	}

	ydbStreamDefaultCodecs = []Ydb_PersQueue_V1.Codec{
		Ydb_PersQueue_V1.Codec_CODEC_RAW,
		Ydb_PersQueue_V1.Codec_CODEC_GZIP,
		Ydb_PersQueue_V1.Codec_CODEC_ZSTD,
	}

	ydbStreamCodecNameToCodec = map[string]Ydb_PersQueue_V1.Codec{
		ydbStreamCodecRAW:  Ydb_PersQueue_V1.Codec_CODEC_RAW,
		ydbStreamCodecGZIP: Ydb_PersQueue_V1.Codec_CODEC_GZIP,
		ydbStreamCodecZSTD: Ydb_PersQueue_V1.Codec_CODEC_ZSTD,
	}

	ydbStreamCodecToCodecName = map[Ydb_PersQueue_V1.Codec]string{
		Ydb_PersQueue_V1.Codec_CODEC_RAW:  ydbStreamCodecRAW,
		Ydb_PersQueue_V1.Codec_CODEC_GZIP: ydbStreamCodecGZIP,
		Ydb_PersQueue_V1.Codec_CODEC_ZSTD: ydbStreamCodecZSTD,
	}
)

func createYDBStreamClient(ctx context.Context, databaseEndpoint string, config *Config) (controlplane.ControlPlane, error) {
	endpoint, databasePath, useTLS, err := parseYandexYDBDatabaseEndpoint(databaseEndpoint)
	if err != nil {
		return nil, err
	}

	opts := session.Options{
		Credentials: credentials.NewAccessTokenCredentials(config.Token),
		Endpoint:    endpoint,
		Database:    databasePath,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{}
	}

	return controlplane.NewControlPlaneClient(ctx, opts)
}

func resourceYDBStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client, err := createYDBStreamClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	var supportedCodecs []Ydb_PersQueue_V1.Codec
	if gotCodecs, ok := d.GetOk("supported_codecs"); !ok {
		supportedCodecs = ydbStreamDefaultCodecs
	} else {
		for _, c := range gotCodecs.([]interface{}) {
			cod := c.(string)
			supportedCodecs = append(supportedCodecs, ydbStreamCodecNameToCodec[cod])
		}
	}

	err = client.CreateTopic(ctx, &Ydb_PersQueue_V1.CreateTopicRequest{
		Path:            d.Get("stream_name").(string),
		OperationParams: &Ydb_Operations.OperationParams{},
		Settings: &Ydb_PersQueue_V1.TopicSettings{
			SupportedCodecs:   supportedCodecs,
			PartitionsCount:   int32(d.Get("partitions_count").(int)),
			RetentionPeriodMs: int64(d.Get("retention_period_ms").(int)),
			SupportedFormat:   Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
		},
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}

	d.SetId(d.Get("database_endpoint").(string) + "/" + d.Get("stream_name").(string))

	return nil
}

func flattenYDBStreamDescription(d *schema.ResourceData, desc *Ydb_PersQueue_V1.DescribeTopicResult) error {
	_ = d.Set("stream_name", desc.Self.Name)
	_ = d.Set("partitions_count", desc.Settings.PartitionsCount)
	_ = d.Set("retention_period_ms", desc.Settings.RetentionPeriodMs)

	supportedCodecs := make([]string, 0, len(desc.Settings.SupportedCodecs))
	for _, v := range desc.Settings.SupportedCodecs {
		switch v {
		case Ydb_PersQueue_V1.Codec_CODEC_RAW:
			supportedCodecs = append(supportedCodecs, ydbStreamCodecRAW)
		case Ydb_PersQueue_V1.Codec_CODEC_ZSTD:
			supportedCodecs = append(supportedCodecs, ydbStreamCodecZSTD)
		case Ydb_PersQueue_V1.Codec_CODEC_GZIP:
			supportedCodecs = append(supportedCodecs, ydbStreamCodecGZIP)
		}
	}

	rules := make([]map[string]interface{}, 0, len(desc.Settings.ReadRules))
	for _, r := range desc.Settings.ReadRules {
		var codecs []string
		for _, codec := range r.SupportedCodecs {
			if c, ok := ydbStreamCodecToCodecName[codec]; ok {
				codecs = append(codecs, c)
			}
		}
		rules = append(rules, map[string]interface{}{
			"name":                          r.ConsumerName,
			"starting_message_timestamp_ms": r.StartingMessageTimestampMs,
			"supported_codecs":              codecs,
			"service_type":                  r.ServiceType,
		})
	}

	err := d.Set("consumers", rules)
	if err != nil {
		return fmt.Errorf("failed to set consumers %+v: %s", rules, err)
	}

	err = d.Set("supported_codecs", supportedCodecs)
	if err != nil {
		return err
	}

	return d.Set("database_endpoint", d.Get("database_endpoint").(string)) // TODO(shmel1k@): remove probably.
}

func resourceYDBStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	client, err := createYDBStreamClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	description, err := client.DescribeTopic(ctx, d.Get("stream_name").(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			d.SetId("") // NOTE(shmel1k@): marking as non-existing resource.
			return nil
		}
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "resource: failed to describe stream",
				Detail:   err.Error(),
			},
		}
	}

	err = flattenYDBStreamDescription(d, description)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to flatten stream description",
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func mergeYDBStreamConsumerSettings(
	consumers []interface{},
	readRules []*Ydb_PersQueue_V1.TopicSettings_ReadRule,
) (newReadRules []*Ydb_PersQueue_V1.TopicSettings_ReadRule) {
	// TODO(shmel1k@): tests.
	rules := make(map[string]*Ydb_PersQueue_V1.TopicSettings_ReadRule, len(readRules))
	for i := 0; i < len(readRules); i++ {
		rules[readRules[i].ConsumerName] = readRules[i]
	}

	if len(consumers) == 0 {
		return readRules
	}

	consumersMap := make(map[string]struct{})
	for _, v := range consumers {
		consumer := v.(map[string]interface{})
		// TODO(shmel1k@): think about fields to add.
		consumerName, ok := consumer["name"].(string)
		if !ok {
			// TODO(shmel1k@): think about error.
			continue
		}

		consumersMap[consumerName] = struct{}{}

		supportedCodecs, ok := consumer["supported_codecs"].([]interface{})
		if !ok {
			for _, vv := range ydbStreamAllowedCodecs {
				supportedCodecs = append(supportedCodecs, vv)
			}
		}
		startingMessageTs, ok := consumer["starting_message_timestamp_ms"].(int)
		if !ok {
			startingMessageTs = 0
		}
		serviceType, ok := consumer["service_type"].(string)
		if !ok {
			serviceType = ""
		}

		r, ok := rules[consumerName]
		if !ok {
			// NOTE(shmel1k@): stream was deleted by someone outside terraform or does not exist.
			codecs := make([]Ydb_PersQueue_V1.Codec, 0, len(supportedCodecs))
			for _, c := range supportedCodecs {
				codec := c.(string)
				codecs = append(codecs, ydbStreamCodecNameToCodec[strings.ToLower(codec)])
			}
			newReadRules = append(newReadRules, &Ydb_PersQueue_V1.TopicSettings_ReadRule{
				ConsumerName:               consumerName,
				SupportedFormat:            Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
				ServiceType:                serviceType,
				StartingMessageTimestampMs: int64(startingMessageTs),
				SupportedCodecs:            codecs,
			})
			continue
		}

		if r.ServiceType != serviceType {
			r.ServiceType = serviceType
		}
		if r.StartingMessageTimestampMs != int64(startingMessageTs) {
			r.StartingMessageTimestampMs = int64(startingMessageTs)
		}

		newCodecs := make([]Ydb_PersQueue_V1.Codec, 0, len(supportedCodecs))
		for _, codec := range supportedCodecs {
			c := ydbStreamCodecNameToCodec[strings.ToLower(codec.(string))]
			newCodecs = append(newCodecs, c)
		}
		if len(newCodecs) != 0 {
			r.SupportedCodecs = newCodecs
		}
		newReadRules = append(newReadRules, r)
	}
	return
}

func mergeYDBStreamSettings(
	d *schema.ResourceData,
	settings *Ydb_PersQueue_V1.TopicSettings,
) *Ydb_PersQueue_V1.TopicSettings {
	if d.HasChange("partitions_count") {
		settings.PartitionsCount = int32(d.Get("partitions_count").(int))
	}
	if d.HasChange("supported_codecs") {
		codecs := d.Get("supported_codecs").([]interface{})
		updatedCodecs := make([]Ydb_PersQueue_V1.Codec, 0, len(codecs))

		for _, c := range codecs {
			cc, ok := ydbStreamCodecNameToCodec[strings.ToLower(c.(string))]
			if !ok {
				// TODO(shmel1k@): add validation of unsupported codecs. Use default if unknown is found.
				panic(fmt.Sprintf("Unsupported codec %q found after validation", cc))
			}
			updatedCodecs = append(updatedCodecs, cc)
		}
		settings.SupportedCodecs = updatedCodecs
	}
	if d.HasChange("retention_period_ms") {
		settings.RetentionPeriodMs = int64(d.Get("retention_period_ms").(int))
	}

	if d.HasChange("consumers") {
		settings.ReadRules = mergeYDBStreamConsumerSettings(d.Get("consumers").([]interface{}), settings.ReadRules)
	}

	return settings
}

func performYandexYDBStreamUpdate(ctx context.Context, d *schema.ResourceData, config *Config) diag.Diagnostics {
	client, err := createYDBStreamClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	streamName := d.Get("stream_name").(string)
	desc, err := client.DescribeTopic(ctx, streamName)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failed to get description for stream %q", streamName),
				Detail:   err.Error(),
			},
		}
	}

	newSettings := mergeYDBStreamSettings(d, desc.GetSettings())

	err = client.AlterTopic(ctx, &Ydb_PersQueue_V1.AlterTopicRequest{
		Path:     streamName,
		Settings: newSettings,
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "got error when tried to alter stream",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceYandexYDBStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	return performYandexYDBStreamUpdate(ctx, d, config)
}

func resourceYandexYDBStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client, err := createYDBStreamClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	streamName := d.Get("stream_name").(string)
	err = client.DropTopic(ctx, &Ydb_PersQueue_V1.DropTopicRequest{
		Path: streamName,
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to delete stream",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceYandexYDBStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYDBStreamCreate,
		ReadContext:   resourceYDBStreamRead,
		UpdateContext: resourceYandexYDBStreamUpdate,
		DeleteContext: resourceYandexYDBStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			// TODO(shmel1k@): think about own timeouts.
			Default: schema.DefaultTimeout(yandexYDBServerlessDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partitions_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  ydbStreamDefaultPartitionsCount,
			},
			"supported_codecs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ydbStreamAllowedCodecs, false),
				},
			},
			"retention_period_ms": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  ydbStreamDefaultRetentionPeriod,
			},
			"consumers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"supported_codecs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(ydbStreamAllowedCodecs, false),
							},
						},
						"starting_message_timestamp_ms": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
