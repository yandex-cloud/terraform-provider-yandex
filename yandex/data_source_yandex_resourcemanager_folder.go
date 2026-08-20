package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexResourceManagerFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a Yandex Resource Manager Folder. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#folder).\n\n~> Either `folder_id` or `name` must be specified.\n",

		Read: dataSourceYandexResourceManagerFolderRead,
		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:        schema.TypeString,
				Description: "ID of the folder.",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"cloud_id": {
				Type:        schema.TypeString,
				Description: "Cloud that the resource belongs to. If value is omitted, the default provider cloud is used.",
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
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Current status of the folder.",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexResourceManagerFolderRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "folder_id", "name")
	if err != nil {
		return err
	}

	folderID := d.Get("folder_id").(string)
	folderName, folderNameOk := d.GetOk("name")

	if folderNameOk {
		dsCloudID, err := getCloudID(d, config)
		if err != nil {
			return fmt.Errorf("error getting cloud ID to resolve data source for folder: %s", err)
		}

		folderID, err = resolveFolderIDByName(ctx, config, folderName.(string), dsCloudID)
		if err != nil {
			return fmt.Errorf("failed to resolve data source folder by name: %v", err)
		}
	}

	client := resourcemanagersdk.NewFolderClient(config.SDK)

	folder, err := client.Get(ctx, &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})

	if err != nil {
		return fmt.Errorf("failed to resolve data source folder by ID: %v", err)
	}

	d.Set("folder_id", folder.Id)
	d.Set("name", folder.Name)
	d.Set("description", folder.Description)
	d.Set("cloud_id", folder.CloudId)
	d.Set("status", strings.ToLower(folder.Status.String()))
	d.Set("created_at", getTimestamp(folder.CreatedAt))

	if err := d.Set("labels", folder.Labels); err != nil {
		return err
	}

	d.SetId(folder.Id)

	return nil
}

func resolveFolderIDByName(ctx context.Context, config *Config, folderName, cloudID string) (string, error) {
	client := resourcemanagersdk.NewFolderClient(config.SDK)

	resolver := resourcemanagersdk.FolderResolver(folderName, client, sdkresolversv2.CloudID(cloudID))
	if err := resolver.Run(ctx); err != nil {
		return "", err
	}
	if err := resolver.Err(); err != nil {
		return "", err
	}

	return resolver.ID(), nil
}
