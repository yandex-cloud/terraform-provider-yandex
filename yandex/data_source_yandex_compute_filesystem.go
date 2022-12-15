package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputeFilesystem() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexComputeFilesystemRead,
		Schema: map[string]*schema.Schema{
			"filesystem_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"block_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexComputeFilesystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "filesystem_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	fsID := d.Get("filesystem_id").(string)
	_, fsNameOk := d.GetOk("name")

	if fsNameOk {
		if fsID, err = resolveObjectID(ctx, config, d, sdkresolvers.FilesystemResolver); err != nil {
			return diag.FromErr(err)
		}
	}

	fs, err := config.sdk.Compute().Filesystem().Get(ctx, &compute.GetFilesystemRequest{
		FilesystemId: fsID,
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("filesystem with ID %q", fsID)))
	}

	d.Set("filesystem_id", fs.Id)
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

	d.SetId(fs.Id)

	return nil
}
