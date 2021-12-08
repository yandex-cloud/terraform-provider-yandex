package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputeImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"family": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"min_disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pooled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"product_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"status": {
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

func dataSourceYandexComputeImageRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	var image *compute.Image

	err := checkOneOf(d, "name", "image_id", "family")
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("family"); ok {
		familyName := v.(string)

		folderID := StandardImagesFolderID
		if f, ok := d.GetOk("folder_id"); ok {
			folderID = f.(string)
		}

		image, err = config.sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
			FolderId: folderID,
			Family:   familyName,
		})

		if err != nil {
			return fmt.Errorf("failed to find latest image with family \"%s\": %s", familyName, err)
		}
	} else {
		imageID := d.Get("image_id").(string)
		_, imageNameOk := d.GetOk("name")

		if imageNameOk {
			imageID, err = resolveObjectID(ctx, config, d, sdkresolvers.ImageResolver)
			if err != nil {
				return fmt.Errorf("failed to resolve data source image by name: %v", err)
			}
		}

		image, err = config.sdk.Compute().Image().Get(ctx, &compute.GetImageRequest{
			ImageId: imageID,
		})

		if err != nil {
			return handleNotFoundError(err, d, fmt.Sprintf("image with ID %q", imageID))
		}
	}

	d.Set("image_id", image.Id)
	d.Set("created_at", getTimestamp(image.CreatedAt))
	d.Set("family", image.Family)
	d.Set("folder_id", image.FolderId)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("status", strings.ToLower(image.Status.String()))
	d.Set("os_type", strings.ToLower(image.Os.Type.String()))
	d.Set("min_disk_size", toGigabytes(image.MinDiskSize))
	d.Set("size", toGigabytes(image.StorageSize))
	d.Set("pooled", image.Pooled)

	if err := d.Set("labels", image.Labels); err != nil {
		return err
	}

	if err := d.Set("product_ids", image.ProductIds); err != nil {
		return err
	}

	d.SetId(image.Id)

	return nil
}
