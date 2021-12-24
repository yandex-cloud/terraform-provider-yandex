package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func dataSourceYandexCDNResource() *schema.Resource {
	resourceSchema := defineYandexCDNResourceBaseSchema()

	resourceSchema.Read = dataSourceYandexCDNResourceRead
	resourceSchema.Schema["resource_id"] = &schema.Schema{
		Type: schema.TypeString,

		Computed: true,
		Optional: true,
	}

	return resourceSchema
}

func resolveCDNResourceID(ctx context.Context, config *Config, d *schema.ResourceData) (string, error) {
	cname := d.Get("cname").(string)

	if cname == "" {
		return "", fmt.Errorf("empty cname for resource")
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return "", err
	}

	iterator := config.sdk.CDN().Resource().ResourceIterator(ctx, &cdn.ListResourcesRequest{
		FolderId: folderID,
	})

	for iterator.Next() {
		cdnResource := iterator.Value()
		if cname == cdnResource.Cname {
			return cdnResource.Id, nil
		}
	}

	return "", fmt.Errorf("resource with cname %q not found", cname)
}

func dataSourceYandexCDNResourceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "resource_id", "cname")
	if err != nil {
		return err
	}

	resourceID := d.Get("resource_id").(string)
	_, resourceCNameOk := d.GetOk("cname")

	if resourceCNameOk {
		resourceID, err = resolveCDNResourceID(ctx, config, d)
		if err != nil {
			return fmt.Errorf("failed to resolve data source cdn resource by name: %v", err)
		}
	}

	resource, err := config.sdk.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: resourceID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("cdn resource with ID %q", resourceID))
	}

	if err := flattenYandexCDNResource(d, resource); err != nil {
		return err
	}

	d.Set("resource_id", resource.Id)
	d.SetId(resource.Id)

	return nil
}
