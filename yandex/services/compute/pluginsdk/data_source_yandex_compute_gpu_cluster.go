package pluginsdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func DataSourceYandexComputeGpuCluster() *schema.Resource {
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
				Description: "Type of interconnect between nodes to use in GPU cluster. Type `INFINIBAND` is set by default, and it is the only one available at the moment.",
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the GPU cluster.",
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexComputeGpuClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*providerconfig.Config)

	err := sdkcommon.CheckOneOf(d, "gpu_cluster_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	gpuClusterID := d.Get("gpu_cluster_id").(string)
	_, gpuClusterNameOk := d.GetOk("name")

	if gpuClusterNameOk {
		if gpuClusterID, err = sdkcommon.ResolveObjectIDV2(ctx, config, d, sdkcommon.ResolverWithClient(computesdk.NewGpuClusterClient(config.SDK), computesdk.GpuClusterResolver)); err != nil {
			return diag.FromErr(err)
		}
	}

	gpuCluster, err := computesdk.NewGpuClusterClient(config.SDK).Get(ctx, &compute.GetGpuClusterRequest{
		GpuClusterId: gpuClusterID,
	})
	if err != nil {
		return diag.FromErr(sdkcommon.HandleNotFoundError(err, d, fmt.Sprintf("GPU cluster with ID %q", gpuClusterID)))
	}

	d.Set("gpu_cluster_id", gpuCluster.Id)
	d.Set("folder_id", gpuCluster.FolderId)
	d.Set("created_at", sdkcommon.GetTimestamp(gpuCluster.CreatedAt))
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
