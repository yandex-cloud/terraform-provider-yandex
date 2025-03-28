package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexComputeGpuCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Compute GPU cluster. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/gpu-cluster).\n\n~> One of `gpu_cluster_id` or `name` should be specified.\n",

		ReadContext: dataSourceYandexComputeGpuClusterRead,
		Schema: map[string]*schema.Schema{
			"gpu_cluster_id": {
				Type:        schema.TypeString,
				Description: "ID of the GPU cluster.",
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"interconnect_type": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeGpuCluster().Schema["interconnect_type"].Description,
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeGpuCluster().Schema["status"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexComputeGpuClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "gpu_cluster_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	gpuClusterID := d.Get("gpu_cluster_id").(string)
	_, gpuClusterNameOk := d.GetOk("name")

	if gpuClusterNameOk {
		if gpuClusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.GpuClusterResolver); err != nil {
			return diag.FromErr(err)
		}
	}

	gpuCluster, err := config.sdk.Compute().GpuCluster().Get(ctx, &compute.GetGpuClusterRequest{
		GpuClusterId: gpuClusterID,
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("GPU cluster with ID %q", gpuClusterID)))
	}

	d.Set("gpu_cluster_id", gpuCluster.Id)
	d.Set("folder_id", gpuCluster.FolderId)
	d.Set("created_at", getTimestamp(gpuCluster.CreatedAt))
	d.Set("name", gpuCluster.Name)
	d.Set("description", gpuCluster.Description)
	d.Set("interconnect_type", strings.ToLower(gpuCluster.InterconnectType.String()))
	d.Set("zone", gpuCluster.ZoneId)
	d.Set("status", strings.ToLower(gpuCluster.Status.String()))

	if err := d.Set("labels", gpuCluster.Labels); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(gpuCluster.Id)

	return nil
}
