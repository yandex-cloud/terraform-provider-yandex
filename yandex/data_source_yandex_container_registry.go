package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexContainerRegistry() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexContainerRegistryRead,
		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceYandexContainerRegistryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "registry_id", "name")
	if err != nil {
		return err
	}

	registryID := d.Get("registry_id").(string)
	_, registryNameOk := d.GetOk("name")

	if registryNameOk {
		registryID, err = resolveObjectID(ctx, config, d, sdkresolvers.RegistryResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Container Registry by name: %v", err)
		}
	}

	registry, err := config.sdk.ContainerRegistry().Registry().Get(ctx,
		&containerregistry.GetRegistryRequest{
			RegistryId: registryID,
		})

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("Сontainer Registry not found: %s", registryID)
		}
		return err
	}

	d.Set("registry_id", registry.Id)
	d.Set("folder_id", registry.FolderId)
	d.Set("name", registry.Name)
	d.Set("status", strings.ToLower(registry.Status.String()))
	d.Set("created_at", getTimestamp(registry.CreatedAt))
	if err := d.Set("labels", registry.Labels); err != nil {
		return err
	}

	d.SetId(registry.Id)

	return nil
}
