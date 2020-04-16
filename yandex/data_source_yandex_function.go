package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexFunctionRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"function_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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

			"execution_timeout": {
				Type:     schema.TypeString,
				Computed: true,
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

			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"loggroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexFunctionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "function_id", "name")
	if err != nil {
		return err
	}

	functionID := d.Get("function_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		functionID, err = resolveObjectID(ctx, config, d, sdkresolvers.FunctionResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud Function by name: %v", err)
		}
	}

	req := functions.GetFunctionRequest{
		FunctionId: functionID,
	}

	function, err := config.sdk.Serverless().Functions().Function().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id()))
	}

	versionReq := functions.GetFunctionVersionByTagRequest{
		FunctionId: function.Id,
		Tag:        "$latest",
	}

	version, err := config.sdk.Serverless().Functions().Function().GetVersionByTag(ctx, &versionReq)
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}

	d.SetId(function.Id)
	d.Set("function_id", function.Id)
	return flattenYandexFunction(d, function, version)
}
