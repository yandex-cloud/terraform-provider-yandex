package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeGpuClusterDefaultTimeout = 5 * time.Minute

func resourceYandexComputeGpuCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexComputeGpuClusterCreate,
		ReadContext:   resourceYandexComputeGpuClusterRead,
		UpdateContext: resourceYandexComputeGpuClusterUpdate,
		DeleteContext: resourceYandexComputeGpuClusterDelete,

		SchemaVersion: 0,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeGpuClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeGpuClusterDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeGpuClusterDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"interconnect_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "INFINIBAND",
				ForceNew: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexComputeGpuClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	zone, err := getZone(d, config)
	if err != nil {
		return diag.Errorf("Error getting zone while creating GPU cluster: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("Error getting folder ID while creating GPU cluster: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating GPU cluster: %s", err)
	}

	interconnectTypeName := strings.ToUpper(d.Get("interconnect_type").(string))
	interconnectType := compute.GpuInterconnectType(compute.GpuInterconnectType_value[interconnectTypeName])

	req := compute.CreateGpuClusterRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		Labels:           labels,
		ZoneId:           zone,
		InterconnectType: interconnectType,
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().GpuCluster().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API for create GPU cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get GPU cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateGpuClusterMetadata)
	if !ok {
		return diag.Errorf("could not get GPU cluster ID from create operation metadata")
	}

	d.SetId(md.GetGpuClusterId())

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while waiting operation to create GPU cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("GPU cluster creation failed: %s", err)
	}

	return resourceYandexComputeGpuClusterRead(ctx, d, meta)
}

func resourceYandexComputeGpuClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	gpuCluster, err := config.sdk.Compute().GpuCluster().Get(ctx, &compute.GetGpuClusterRequest{
		GpuClusterId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("GPU cluster %q", d.Id())))
	}

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

	return nil
}

func resourceYandexComputeGpuClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var resourceComputeGpuClusterUpdateFieldsMap = map[string]string{
		"name":        "name",
		"description": "description",
		"labels":      "labels",
	}

	d.Partial(true)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := compute.UpdateGpuClusterRequest{
		GpuClusterId: d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
	}

	paths := generateFieldMasks(d, resourceComputeGpuClusterUpdateFieldsMap)
	if len(paths) > 0 {
		req.UpdateMask = &fieldmaskpb.FieldMask{Paths: paths}
		if err := updateGpuCluster(ctx, &req, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)

	return resourceYandexComputeGpuClusterRead(ctx, d, meta)
}

func resourceYandexComputeGpuClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().GpuCluster().Delete(
		ctx, &compute.DeleteGpuClusterRequest{
			GpuClusterId: d.Id(),
		}))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("GpuCluster %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateGpuCluster(ctx context.Context, req *compute.UpdateGpuClusterRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().GpuCluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update GPU cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating GPU cluster %q: %s", d.Id(), err)
	}

	return nil
}
