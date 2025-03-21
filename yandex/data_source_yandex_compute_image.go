package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexComputeImage() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Compute image. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/image).\n\n~> Either `image_id`, `family` or `name` must be specified.\n\n~> If you specify `family` without `folder_id` then lookup takes place in the 'standard-images' folder.\n",

		Read: dataSourceYandexComputeImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"image_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific image.",
				Optional:    true,
				Computed:    true,
			},
			"family": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeImage().Schema["family"].Description,
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"min_disk_size": {
				Type:        schema.TypeInt,
				Description: resourceYandexComputeImage().Schema["min_disk_size"].Description,
				Computed:    true,
			},
			"size": {
				Type:        schema.TypeInt,
				Description: resourceYandexComputeImage().Schema["size"].Description,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"os_type": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeImage().Schema["os_type"].Description,
				Computed:    true,
			},
			"pooled": {
				Type:        schema.TypeBool,
				Description: resourceYandexComputeImage().Schema["pooled"].Description,
				Computed:    true,
			},
			"product_ids": {
				Type:        schema.TypeSet,
				Description: resourceYandexComputeImage().Schema["product_ids"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeImage().Schema["status"].Description,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"hardware_generation": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"legacy_features": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pci_topology": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
							Computed: true,
						},

						"generation2_features": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"kms_key_id": {
				Type:        schema.TypeString,
				Description: "ID of KMS symmetric key used to encrypt image.",
				Computed:    true,
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

	hardwareGeneration, err := flattenComputeHardwareGeneration(image.HardwareGeneration)
	if err != nil {
		return err
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

	if image.KmsKey != nil {
		d.Set("kms_key_id", image.KmsKey.KeyId)
	}

	if err := d.Set("labels", image.Labels); err != nil {
		return err
	}

	if err := d.Set("product_ids", image.ProductIds); err != nil {
		return err
	}

	if err := d.Set("hardware_generation", hardwareGeneration); err != nil {
		return err
	}

	d.SetId(image.Id)

	return nil
}
