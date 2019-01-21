package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

const yandexResourceManagerFolderDefaultTimeout = 1 * time.Minute

func dataSourceYandexResourceManagerFolder() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexResourceManagerFolderRead,
		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"cloud_id": {
				Type:     schema.TypeString,
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

func dataSourceYandexResourceManagerFolderRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folder, err := config.sdk.ResourceManager().Folder().Get(context.Background(), &resourcemanager.GetFolderRequest{
		FolderId: d.Get("folder_id").(string),
	})

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("folder not found: %s", d.Get("folder_id").(string))
		}

		return err
	}

	createdAt, err := getTimestamp(folder.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("name", folder.Name)
	d.Set("description", folder.Description)
	d.Set("cloud_id", folder.CloudId)
	d.Set("status", strings.ToLower(folder.Status.String()))
	d.Set("created_at", createdAt)

	if err := d.Set("labels", folder.Labels); err != nil {
		return err
	}

	d.SetId(folder.Id)

	return nil
}
