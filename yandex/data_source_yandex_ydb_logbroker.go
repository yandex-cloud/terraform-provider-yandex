package yandex

import (
	"context"
	"log"

	"github.com/ydb-platform/ydb-go-persqueue-sdk/session"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-persqueue-sdk/controlplane"
)

func dataSourceYandexYDSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println(d.Get("token"))
	client, err := controlplane.NewControlPlaneClient(ctx, session.Options{
		Credentials: credentials.NewAccessTokenCredentials(d.Get("token").(string)),
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds-controlplane client",
				Detail:   err.Error(),
			},
		}
	}
	_ = client

	return nil
}

func dataSourceYandexYDSServerless() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexYDSRead,

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
