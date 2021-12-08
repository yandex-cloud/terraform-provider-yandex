package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeImageDefaultTimeout = 5 * time.Minute
const StandardImagesFolderID = "standard-images"

func resourceYandexComputeImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeImageCreate,
		Read:   resourceYandexComputeImageRead,
		Update: resourceYandexComputeImageUpdate,
		Delete: resourceYandexComputeImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeImageDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeImageDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeImageDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"family": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"min_disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"pooled": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"source_family": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_snapshot", "source_disk", "source_url", "source_image"},
			},

			"source_image": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_snapshot", "source_disk", "source_url", "source_family"},
			},

			"source_snapshot": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_image", "source_disk", "source_url", "source_family"},
			},

			"source_disk": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_image", "source_snapshot", "source_url", "source_family"},
			},

			"source_url": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_image", "source_snapshot", "source_disk", "source_family"},
			},

			"product_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"size": {
				Type:     schema.TypeInt,
				Computed: true,
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

func resourceYandexComputeImageCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating image: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating image: %s", err)
	}

	productIds, err := expandProductIds(d.Get("product_ids"))
	if err != nil {
		return fmt.Errorf("Error expanding product IDs while creating image: %s", err)
	}

	osTypeName := strings.ToUpper(d.Get("os_type").(string))
	osType := compute.Os_Type(compute.Os_Type_value[osTypeName])

	req := compute.CreateImageRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Family:      d.Get("family").(string),
		Labels:      labels,
		MinDiskSize: toBytes(d.Get("min_disk_size").(int)),
		Pooled:      d.Get("pooled").(bool),
		ProductIds:  productIds,
		Os: &compute.Os{
			Type: osType,
		},
	}

	err = prepareSourceForImage(&req, d, meta)
	if err != nil {
		return fmt.Errorf("Error while prepare request to create image: %s", err)
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Image().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create image: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get image create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateImageMetadata)
	if !ok {
		return fmt.Errorf("could not get Image ID from create operation metadata")
	}

	d.SetId(md.ImageId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create image: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Image creation failed: %s", err)
	}

	return resourceYandexComputeImageRead(d, meta)
}

func resourceYandexComputeImageRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	image, err := config.sdk.Compute().Image().Get(config.Context(), &compute.GetImageRequest{
		ImageId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Image %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(image.CreatedAt))
	d.Set("name", image.Name)
	d.Set("folder_id", image.FolderId)
	d.Set("description", image.Description)
	d.Set("min_disk_size", toGigabytes(image.MinDiskSize))
	d.Set("status", strings.ToLower(image.Status.String()))
	d.Set("family", image.Family)
	d.Set("size", toGigabytes(image.StorageSize))
	d.Set("pooled", image.Pooled)

	if err := d.Set("labels", image.Labels); err != nil {
		return err
	}

	return d.Set("product_ids", image.ProductIds)
}

func resourceYandexComputeImageUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	labelPropName := "labels"
	if d.HasChange(labelPropName) {
		labelsProp, err := expandLabels(d.Get(labelPropName))
		if err != nil {
			return err
		}

		req := &compute.UpdateImageRequest{
			ImageId: d.Id(),
			Labels:  labelsProp,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelPropName},
			},
		}

		err = makeImageUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req := &compute.UpdateImageRequest{
			ImageId: d.Id(),
			Name:    d.Get(namePropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{namePropName},
			},
		}

		err := makeImageUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req := &compute.UpdateImageRequest{
			ImageId:     d.Id(),
			Description: d.Get(descPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{descPropName},
			},
		}

		err := makeImageUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	minDiskSizePropName := "min_disk_size"
	if d.HasChange(minDiskSizePropName) {
		req := &compute.UpdateImageRequest{
			ImageId:     d.Id(),
			MinDiskSize: toBytes(d.Get(minDiskSizePropName).(int)),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{minDiskSizePropName},
			},
		}

		err := makeImageUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	d.Partial(false)

	return resourceYandexComputeImageRead(d, meta)
}

func resourceYandexComputeImageDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Image %q", d.Id())

	req := &compute.DeleteImageRequest{
		ImageId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Image().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Image %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Image %q", d.Id())
	return nil
}

func prepareSourceForImage(req *compute.CreateImageRequest, d *schema.ResourceData, meta interface{}) error {
	sourceAttrs := []string{"source_family", "source_disk", "source_image", "source_snapshot", "source_url"}
	var selectedSourceAttr string
	var selectedSourceValue string

	for _, attrName := range sourceAttrs {
		if v, ok := d.GetOk(attrName); ok {
			if selectedSourceAttr == "" {
				selectedSourceAttr = attrName
				selectedSourceValue = v.(string)
			} else {
				return fmt.Errorf("more than one source attribute present: %s and %s, only one allowed", selectedSourceAttr, attrName)
			}

		}
	}

	switch selectedSourceAttr {
	case "source_family":
		config := meta.(*Config)
		ctx := config.Context()
		familyName := d.Get("source_family").(string)
		img, err := config.sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
			FolderId: StandardImagesFolderID,
			Family:   familyName,
		})
		if err != nil {
			return fmt.Errorf("failed to find image with family \"%s\": %s", familyName, err)
		}
		req.Source = &compute.CreateImageRequest_ImageId{
			ImageId: img.Id,
		}
	case "source_disk":
		req.Source = &compute.CreateImageRequest_DiskId{
			DiskId: selectedSourceValue,
		}
	case "source_image":
		req.Source = &compute.CreateImageRequest_ImageId{
			ImageId: selectedSourceValue,
		}
	case "source_snapshot":
		req.Source = &compute.CreateImageRequest_SnapshotId{
			SnapshotId: selectedSourceValue,
		}
	case "source_url":
		req.Source = &compute.CreateImageRequest_Uri{
			Uri: selectedSourceValue,
		}
	default:
		// should not occur: validation must be done at Schema level
		return fmt.Errorf("selected source attr %s not one from %s", selectedSourceAttr, sourceAttrs)
	}

	return nil
}

func makeImageUpdateRequest(req *compute.UpdateImageRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Image().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Image %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Image %q: %s", d.Id(), err)
	}

	return nil
}
