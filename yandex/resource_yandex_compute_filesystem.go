package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeFilesystemDefaultTimeout = 5 * time.Minute

func resourceYandexComputeFilesystem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexComputeFilesystemCreate,
		ReadContext:   resourceYandexComputeFilesystemRead,
		UpdateContext: resourceYandexComputeFilesystemUpdate,
		DeleteContext: resourceYandexComputeFilesystemDelete,

		SchemaVersion: 0,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeFilesystemDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeFilesystemDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeFilesystemDefaultTimeout),
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
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "network-hdd",
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      150,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"block_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      4096,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexComputeFilesystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	zone, err := getZone(d, config)
	if err != nil {
		return diag.Errorf("Error getting zone while creating filesystem: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("Error getting folder ID while creating filesystem: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating filesystem: %s", err)
	}

	req := compute.CreateFilesystemRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		TypeId:      d.Get("type").(string),
		ZoneId:      zone,
		Size:        toBytes(d.Get("size").(int)),
		BlockSize:   int64(d.Get("block_size").(int)),
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Filesystem().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API for create filesystem: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get filesystem create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateFilesystemMetadata)
	if !ok {
		return diag.Errorf("could not get filesystem ID from create operation metadata")
	}

	d.SetId(md.GetFilesystemId())

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while waiting operation to create filesystem: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("Filesystem creation failed: %s", err)
	}

	return resourceYandexComputeFilesystemRead(ctx, d, meta)
}

func resourceYandexComputeFilesystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	fs, err := config.sdk.Compute().Filesystem().Get(ctx, &compute.GetFilesystemRequest{
		FilesystemId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("filesystem %q", d.Id())))
	}

	d.Set("folder_id", fs.FolderId)
	d.Set("created_at", getTimestamp(fs.CreatedAt))
	d.Set("name", fs.Name)
	d.Set("description", fs.Description)
	d.Set("type", fs.TypeId)
	d.Set("zone", fs.ZoneId)
	d.Set("size", toGigabytes(fs.Size))
	d.Set("block_size", int(fs.BlockSize))
	d.Set("status", strings.ToLower(fs.Status.String()))

	if err := d.Set("labels", fs.Labels); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexComputeFilesystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var resourceComputeFilesystemUpdateFieldsMap = map[string]string{
		"name":        "name",
		"description": "description",
		"labels":      "labels",
	}

	d.Partial(true)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := compute.UpdateFilesystemRequest{
		FilesystemId: d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		Size:         toBytes(d.Get("size").(int)),
	}

	if d.HasChange("size") {
		req.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"size"}}
		if err := updateFilesystem(ctx, &req, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	paths := generateFieldMasks(d, resourceComputeFilesystemUpdateFieldsMap)
	if len(paths) > 0 {
		req.UpdateMask = &fieldmaskpb.FieldMask{Paths: paths}
		if err := updateFilesystem(ctx, &req, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)

	return resourceYandexComputeFilesystemRead(ctx, d, meta)
}

func resourceYandexComputeFilesystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Filesystem().Delete(
		ctx, &compute.DeleteFilesystemRequest{
			FilesystemId: d.Id(),
		}))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Filesystem %q", d.Id())))
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

func updateFilesystem(ctx context.Context, req *compute.UpdateFilesystemRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Filesystem().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update filesystem %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating filesystem %q: %s", d.Id(), err)
	}

	return nil
}
