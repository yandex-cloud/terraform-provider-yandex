package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

func dataSourceYandexFunctionVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexFunctionVersionRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"entrypoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"concurrency": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"environment": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"image_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"loggroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secrets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"environment_variable": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexFunctionVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := checkOneOf(d, "function_id", "version_id"); err != nil {
		return diag.FromErr(err)
	}

	var (
		version *functions.Version
		err     error
	)

	if functionID, ok := d.GetOk("function_id"); ok {
		// Resolve by function id and tag.
		tag, ok := d.GetOk("tag")
		if !ok {
			return diag.Errorf("failed to resolve Yandex Cloud Function: tag must be set for function_id")
		}
		req := &functions.GetFunctionVersionByTagRequest{
			FunctionId: functionID.(string),
			Tag:        tag.(string),
		}

		version, err = config.sdk.Serverless().Functions().Function().GetVersionByTag(ctx, req)
		if err != nil {
			return handleNotFoundDiagError(err, d, fmt.Sprintf("Yandex Cloud Function Version (function_id, tag) (%q, %q)", functionID.(string), tag.(string)))
		}
	} else if versionID, ok := d.GetOk("version_id"); ok {
		// Resolve by version id.
		req := &functions.GetFunctionVersionRequest{
			FunctionVersionId: versionID.(string),
		}
		version, err = config.sdk.Serverless().Functions().Function().GetVersion(ctx, req)
		if err != nil {
			return handleNotFoundDiagError(err, d, fmt.Sprintf("version %q", versionID.(string)))
		}
	} else {
		return diag.Errorf("failed to resolve data source Yandex Cloud Function: any of `function_id, tag` or `version_id` must be set")
	}

	d.SetId(version.Id)
	if err = d.Set("function_id", version.FunctionId); err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(flattenYandexFunctionVersion(d, version))
}
