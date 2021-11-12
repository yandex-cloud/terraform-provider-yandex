package yandex

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexClientConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexClientConfigRead,
		Schema: map[string]*schema.Schema{
			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceYandexClientConfigRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	response, err := config.sdk.CreateIAMToken(ctx)
	if err != nil {
		return err
	}

	iamToken := response.GetIamToken()
	d.Set("cloud_id", config.CloudID)
	d.Set("folder_id", config.FolderID)
	d.Set("zone", config.Zone)
	d.Set("iam_token", iamToken)
	d.SetId(strconv.Itoa(schema.HashString(fmt.Sprintf("%s:%s:%s:%s", config.CloudID, config.FolderID, config.Zone, iamToken))))

	return nil
}
