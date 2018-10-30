package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

func dataSourceYandexResourceManagerCloud() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexResourceManagerCloudRead,
		Schema: map[string]*schema.Schema{
			"cloud_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"cloud_id"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexResourceManagerCloudRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var cloud *resourcemanager.Cloud
	if v, ok := d.GetOk("cloud_id"); ok {
		resp, err := config.sdk.ResourceManager().Cloud().Get(context.Background(), &resourcemanager.GetCloudRequest{
			CloudId: v.(string),
		})

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				return fmt.Errorf("cloud not found: %s", v)
			}
			return err
		}

		cloud = resp
	} else if v, ok := d.GetOk("name"); ok {
		filter := fmt.Sprintf(`name = "%s"`, v.(string))
		resp, err := config.sdk.ResourceManager().Cloud().List(context.Background(), &resourcemanager.ListCloudsRequest{
			Filter: filter,
		})
		if err != nil {
			return fmt.Errorf("error reading cloud: %s", err)
		}
		if len(resp.Clouds) == 0 {
			return fmt.Errorf("cloud not found: %s", v)
		}

		if len(resp.Clouds) > 1 {
			return fmt.Errorf("more than one matching cloud found")
		}
		cloud = resp.Clouds[0]
	} else {
		return fmt.Errorf("one of 'cloud_id' or 'name' must be set")
	}

	createdAt, err := getTimestamp(cloud.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("name", cloud.Name)
	d.Set("description", cloud.Description)
	d.Set("created_at", createdAt)
	d.SetId(cloud.Id)

	return nil
}
