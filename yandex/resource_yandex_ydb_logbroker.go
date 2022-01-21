package yandex

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_PersQueue_V1"

	"github.com/ydb-platform/ydb-go-persqueue-sdk/controlplane"
	"github.com/ydb-platform/ydb-go-persqueue-sdk/session"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createYDSServerlessClient(ctx context.Context, databaseEndpoint string, config *Config) (controlplane.ControlPlane, error) {
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

func resourceYandexYDSServerlessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client, err := createYDSServerlessClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}

	err = client.CreateTopic(ctx, &Ydb_PersQueue_V1.CreateTopicRequest{
		Path:            d.Get("stream_name").(string),
		OperationParams: &Ydb_Operations.OperationParams{},
		Settings: &Ydb_PersQueue_V1.TopicSettings{
			SupportedCodecs: []Ydb_PersQueue_V1.Codec{
				// TODO(shmel1k@): add mapping.
				Ydb_PersQueue_V1.Codec_CODEC_GZIP,
			},
			PartitionsCount:   2,
			RetentionPeriodMs: 100000000,
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
	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Detail:   fmt.Sprintf("%s", d.Get("database_endpoint")),
			Summary:  "hello, world!",
		},
	}
}

func resourceYandexYDSServerlessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceYandexYDSServerlessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceYandexYDSServerlessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceYandexYDSServerless() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexYDSServerlessCreate,
		ReadContext:   resourceYandexYDSServerlessRead,
		UpdateContext: resourceYandexYDSServerlessUpdate,
		DeleteContext: resourceYandexYDSServerlessDelete,
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
			},
			"supported_codecs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					// TODO(shmel1k@): add validation.
					Type: schema.TypeString,
				},
			},
		},
	}
}
