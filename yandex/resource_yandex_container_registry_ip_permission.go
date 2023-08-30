package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk/operation"
)

const yandexContainerRegistryIPPermissionDefaultTimeout = 5 * time.Minute

func resourceYandexContainerRegistryIPPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexContainerRegistryIPPermissionCreate,
		ReadContext:   resourceYandexContainerRegistryIPPermissionRead,
		UpdateContext: resourceYandexContainerRegistryIPPermissionUpdate,
		DeleteContext: resourceYandexContainerRegistryIPPermissionDelete,

		Importer: &schema.ResourceImporter{
			State: resourceYandexContainerRegistryIPPermissionImporterFunc,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexContainerRegistryIPPermissionDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"push": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Optional: true,

				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrBlocks,
				},

				AtLeastOneOf: []string{"push", "pull"},
			},

			"pull": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Optional: true,

				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrBlocks,
				},

				AtLeastOneOf: []string{"push", "pull"},
			},
		},
	}
}

func resourceYandexContainerRegistryIPPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Set IP Permissions")

	registryId := d.Get("registry_id").(string)
	ipPermissions := append(
		getContainerRegistryIPPermission(d, containerregistry.IpPermission_PULL),
		getContainerRegistryIPPermission(d, containerregistry.IpPermission_PUSH)...)
	log.Printf("[DEBUG] IP Permission to set: %v", stringifyContainerRegistryIPPermission(ipPermissions))

	config := meta.(*Config)
	containerRegistryService := config.sdk.ContainerRegistry().Registry()
	setIPPermissionRequest := containerregistry.SetIpPermissionRequest{
		RegistryId:   registryId,
		IpPermission: ipPermissions,
	}
	_, err := containerRegistryService.SetIpPermission(ctx, &setIPPermissionRequest)
	if err != nil {
		log.Printf("[DEBUG] IP Permissions were not set for Container Registry: %v (err: %s)", registryId, err.Error())
	}

	d.SetId(registryId + containerRegistryIPPermissionIDSuffix)

	log.Printf("[DEBUG] Finished set IP Permissions for Container Registry: %v", registryId)

	return resourceYandexContainerRegistryIPPermissionRead(ctx, d, meta)
}

const containerRegistryIPPermissionIDSuffix = "/ip_permission"

func resourceYandexContainerRegistryIPPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Read IP Permissions")

	config := meta.(*Config)

	registryId := d.Get("registry_id").(string)
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

	d.Set("push", flattenContainerRegistryIPPermissionCIDRs(push))
	d.Set("pull", flattenContainerRegistryIPPermissionCIDRs(pull))

	log.Printf("[DEBUG] Finished read IP Permissions: %v", stringifyContainerRegistryIPPermission(append(push, pull...)))

	return nil
}

func resourceYandexContainerRegistryIPPermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Update IP Permissions")

	if d.HasChange("pull") || d.HasChange("push") {
		registryId := d.Get("registry_id").(string)
		ipPermissions := append(
			getContainerRegistryIPPermission(d, containerregistry.IpPermission_PULL),
			getContainerRegistryIPPermission(d, containerregistry.IpPermission_PUSH)...)
		log.Printf("[DEBUG] IP Permission to set: %v", stringifyContainerRegistryIPPermission(ipPermissions))

		config := meta.(*Config)
		containerRegistryService := config.sdk.ContainerRegistry().Registry()
		ipPermissionUpdateRequest := containerregistry.SetIpPermissionRequest{
			RegistryId:   registryId,
			IpPermission: ipPermissions,
		}
		operation, err := config.sdk.WrapOperation(containerRegistryService.SetIpPermission(ctx, &ipPermissionUpdateRequest))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := waitContainerRegistryIPPermissionOperation(ctx, operation); err != nil {
			return diag.FromErr(err)
		}
	}

	log.Print("[DEBUG] Finished update IP Permissions")

	return nil
}

func resourceYandexContainerRegistryIPPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Deleting Container Registry IP Permissions")

	config := meta.(*Config)
	containerRegistryService := config.sdk.ContainerRegistry().Registry()
	ipPermissionDeleteRequest := containerregistry.SetIpPermissionRequest{
		RegistryId:   d.Get("registry_id").(string),
		IpPermission: []*containerregistry.IpPermission{}, // empty list assumes deletion
	}
	operation, err := config.sdk.WrapOperation(containerRegistryService.SetIpPermission(ctx, &ipPermissionDeleteRequest))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitContainerRegistryIPPermissionOperation(ctx, operation); err != nil {
		return diag.FromErr(err)
	}

	log.Print("[DEBUG] Finished deleting Container Registry IP Permission")

	return nil
}

func resourceYandexContainerRegistryIPPermissionImporterFunc(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// to import permission registry_id must be used
	// because unfortunately this resource don't have own ID

	var (
		passedIdForImport = d.Id()

		parts = strings.Split(passedIdForImport, "/")

		registryId string
	)

	switch len(parts) {
	case 1:
		registryId = parts[0]

	case 2:
		if parts[1] != strings.TrimLeft(containerRegistryIPPermissionIDSuffix, `/`) {
			return nil, fmt.Errorf(`to import resource must be passed "registryId" or "registryId\ip_permission" as value, got: %v`, passedIdForImport)
		}

		registryId = parts[0]
	default:
		return nil, fmt.Errorf(`to import resource must be passed "registryId" or "registryId\ip_permission" as value, got: %v`, passedIdForImport)
	}

	d.Set("registry_id", registryId)
	d.SetId(registryId + containerRegistryIPPermissionIDSuffix)

	return []*schema.ResourceData{d}, nil
}

func waitContainerRegistryIPPermissionOperation(ctx context.Context, operation *operation.Operation) error {
	if err := operation.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to set ip permission: %s", err)
	}

	if _, err := operation.Response(); err != nil {
		return fmt.Errorf("failed to set ip permission: %s", err)
	}

	return nil
}

func getContainerRegistryIPPermission(d *schema.ResourceData, action containerregistry.IpPermission_Action) []*containerregistry.IpPermission {
	var permissions []*containerregistry.IpPermission

	perm := d.Get(strings.ToLower(action.String()))
	expanded := expandContainerRegistryIPPermission(perm, action)
	permissions = append(permissions, expanded...)

	return permissions
}

func expandContainerRegistryIPPermission(v interface{}, action containerregistry.IpPermission_Action) []*containerregistry.IpPermission {
	var (
		perm        = v.(*schema.Set)
		permissions = make([]*containerregistry.IpPermission, perm.Len())
	)

	for i, address := range expandStringSlice(perm.List()) {
		permissions[i] = &containerregistry.IpPermission{
			Ip:     address,
			Action: action,
		}
	}

	return permissions
}

func flattenContainerRegistryIPPermissionCIDRs(ipPermissions []*containerregistry.IpPermission) []string {
	var cidrs []string

	if len(ipPermissions) > 0 {
		for _, perm := range ipPermissions {
			cidrs = append(cidrs, perm.GetIp())
		}
	}

	return cidrs
}

func stringifyContainerRegistryIPPermission(ipPermissions []*containerregistry.IpPermission) string {
	var (
		push []string
		pull []string
		unkn []string
	)

	for _, perm := range ipPermissions {
		switch perm.GetAction() {
		case containerregistry.IpPermission_PUSH:
			push = append(pull, perm.GetIp())
		case containerregistry.IpPermission_PULL:
			pull = append(push, perm.GetIp())
		default:
			unkn = append(unkn, perm.GetIp())
		}
	}

	// unexpected, but we want to know if it happens
	if len(unkn) > 0 {
		return fmt.Sprintf("push: [ %v ], pull [ %v ], unknown: [ %v }",
			stringifyContainerRegistryIPPermissionSlice(push),
			stringifyContainerRegistryIPPermissionSlice(pull),
			stringifyContainerRegistryIPPermissionSlice(unkn))
	}

	return fmt.Sprintf("push: [ %v ], pull [ %v ]",
		stringifyContainerRegistryIPPermissionSlice(push),
		stringifyContainerRegistryIPPermissionSlice(pull))
}

func stringifyContainerRegistryIPPermissionSlice(ipPermissions []string) string {
	return `"` + strings.Join(ipPermissions, `", "`) + `"`
}
