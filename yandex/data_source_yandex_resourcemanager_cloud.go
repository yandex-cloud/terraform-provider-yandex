package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexResourceManagerCloud() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get cloud details. For more information, see [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#cloud).\n\n~> Either `cloud_id` or `name` must be specified.\n",

		Read: dataSourceYandexResourceManagerCloudRead,
		Schema: map[string]*schema.Schema{
			"cloud_id": {
				Type:        schema.TypeString,
				Description: "ID of the cloud.",
				Optional:    true,
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
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
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

	if err != nil {
		return fmt.Errorf("failed to resolve data source cloud by id: %v", err)
	}

	d.Set("cloud_id", cloud.Id)
	d.Set("name", cloud.Name)
	d.Set("description", cloud.Description)
	d.Set("created_at", getTimestamp(cloud.CreatedAt))
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
