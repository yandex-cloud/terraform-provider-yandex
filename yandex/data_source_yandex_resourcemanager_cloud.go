package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexResourceManagerCloud() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexResourceManagerCloudRead,
		Schema: map[string]*schema.Schema{
			"cloud_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folders": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexResourceManagerCloudRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "cloud_id", "name")
	if err != nil {
		return err
	}

	cloudID := d.Get("cloud_id").(string)
	cloudName, cloudNameOk := d.GetOk("name")

	if cloudNameOk {
		cloudID, err = resolveCloudIDByName(ctx, config, cloudName.(string))
		if err != nil {
			return fmt.Errorf("failed to resolve data source cloud by name: %v", err)
		}
	}

	cloud, err := config.sdk.ResourceManager().Cloud().Get(ctx, &resourcemanager.GetCloudRequest{
		CloudId: cloudID,
	})

	req := &resourcemanager.ListFoldersRequest{CloudId: cloudID}
	it := config.sdk.ResourceManager().Folder().FolderIterator(ctx, req)
	var folders []map[string]string
	for it.Next() {
		id := it.Value().GetId()
		name := it.Value().Name
		folders = append(folders, map[string]string{"name": name, "folder_id": id})
	}

	if err != nil {
		return fmt.Errorf("failed to resolve data source cloud by id: %v", err)
	}

	d.Set("cloud_id", cloud.Id)
	d.Set("name", cloud.Name)
	d.Set("description", cloud.Description)
	d.Set("created_at", getTimestamp(cloud.CreatedAt))
	d.Set("folders", folders)
	d.SetId(cloud.Id)

	return nil
}

func resolveCloudIDByName(ctx context.Context, config *Config, name string) (string, error) {
	var objectID string
	resolver := sdkresolvers.CloudResolver(name, sdkresolvers.Out(&objectID))

	err := config.sdk.Resolve(ctx, resolver)
	if err != nil {
		return "", err
	}

	return objectID, nil
}
