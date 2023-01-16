package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexContainerRegistryIPPermission() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexContainerRegistryIPPermissionRead,

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexContainerRegistryDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"registry_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"registry_id"},
			},

			"registry_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"registry_name"},
			},

			"push": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Computed: true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"pull": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Computed: true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceYandexContainerRegistryIPPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	registryId, err := resolveRegistryID(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	config := meta.(*Config)

	containerRegistryService := config.sdk.ContainerRegistry().Registry()
	listIPPermissionRequest := containerregistry.ListIpPermissionRequest{
		RegistryId: registryId,
	}
	listIPPermissionResponse, err := containerRegistryService.ListIpPermission(ctx, &listIPPermissionRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		push []*containerregistry.IpPermission
		pull []*containerregistry.IpPermission
	)

	for _, v := range listIPPermissionResponse.GetPermissions() {
		switch v.Action {
		case containerregistry.IpPermission_PULL:
			pull = append(pull, v)
		case containerregistry.IpPermission_PUSH:
			push = append(push, v)
		}
	}

	log.Printf("[DEBUG] Got Container Registry IP Permissions: %v", stringifyContainerRegistryIPPermission(append(push, pull...)))

	d.Set("push", flattenContainerRegistryIPPermissionCIDRs(push))
	d.Set("pull", flattenContainerRegistryIPPermissionCIDRs(pull))
	d.SetId(registryId + containerRegistryIPPermissionIDSuffix)

	return nil
}

func resolveRegistryID(ctx context.Context, d *schema.ResourceData, meta interface{}) (string, error) {
	err := checkOneOf(d, "registry_id", "registry_name")
	if err != nil {
		return "", err
	}

	config := meta.(*Config)

	var registryID string
	rid, ok := d.GetOk("registry_id")
	if ok {
		registryID = rid.(string)
	} else {
		name, ok := d.GetOk("registry_name")
		if !ok {
			return "", fmt.Errorf("non empty registry_name should be provided")
		}

		folderID, err := getFolderID(d, config)
		if err != nil {
			return "", err
		}

		registryID, err = resolveObjectIDByNameAndFolderID(ctx, config, name.(string), folderID, sdkresolvers.RegistryResolver)
		if err != nil {
			return "", fmt.Errorf("failed to resolve data source Container Registry by name: %v", err)
		}
	}

	return registryID, nil
}
