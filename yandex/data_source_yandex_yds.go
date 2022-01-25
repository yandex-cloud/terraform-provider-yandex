package yandex

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYDBStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			d.SetId("")
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

func dataSourceYandexYDBStream() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYDBStreamRead,

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
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ydbStreamAllowedCodecs, false),
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(ydbStreamAllowedCodecs, false),
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
