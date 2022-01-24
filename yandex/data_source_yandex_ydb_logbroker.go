package yandex

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexYDSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	defer func() {
		_ = client.Close()
	}()

	// TODO(shmel1k@): remove copypaste.
	description, err := client.DescribeTopic(ctx, d.Get("stream_name").(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			d.MarkNewResource()
			// TODO(shmel1k@): handle carefully
			return nil
		}
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "datasource: failed to describe stream",
				Detail:   err.Error(),
			},
		}
	}

	err = flattenYDSDescription(d, description)
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
			"retention_period_ms": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000 * 60 * 60 * 24, // 1 day
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
								// TODO(shmel1k@): add validation.
								Type: schema.TypeString,
							},
						},
						"starting_message_timestamp_ms": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"service_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
