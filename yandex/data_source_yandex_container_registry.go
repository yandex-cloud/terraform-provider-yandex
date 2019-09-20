package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

func dataSourceYandexContainerRegistry() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexContainerRegistryRead,
		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
	ctx := config.ContextWithClientTraceID()
	var registry *containerregistry.Registry

	v, ok := d.GetOk("registry_id")
	if !ok {
		return fmt.Errorf("'registry_id' must be set")
	}

	registry, err := config.sdk.ContainerRegistry().Registry().Get(ctx,
		&containerregistry.GetRegistryRequest{
			RegistryId: v.(string),
		})

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("registry not found: %s", v)
		}
		return err
	}

	createdAt, err := getTimestamp(registry.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("registry_id", registry.Id)
	d.Set("folder_id", registry.FolderId)
	d.Set("name", registry.Name)
	d.Set("status", strings.ToLower(registry.Status.String()))
	d.Set("created_at", createdAt)
	if err := d.Set("labels", registry.Labels); err != nil {
		return err
	}

	d.SetId(registry.Id)

	return nil
}
