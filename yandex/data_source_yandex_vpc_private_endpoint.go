package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	privatelink "github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
	privatelinksdk "github.com/yandex-cloud/go-sdk/services/vpc/v1/privatelink"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexVPCPrivateEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC Private Endpoint. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts/index).\n\nThis data source is used to define [VPC Private Endpoint](https://yandex.cloud/docs/vpc/concepts/private-endpoint) that can be used by other resources.\n\n~> One of `private_endpoint_id` or `name` should be specified.\n",

		Read: dataSourceYandexVPCPrivateEndpointRead,
		Schema: map[string]*schema.Schema{
			"private_endpoint_id": {
				Type:        schema.TypeString,
				Description: "ID of the private endpoint.",
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
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
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCPrivateEndpoint().Schema["status"].Description,
				Computed:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCPrivateEndpoint().Schema["network_id"].Description,
				Computed:    true,
			},
			"object_storage": {
				Type:        schema.TypeList,
				Description: resourceYandexVPCPrivateEndpoint().Schema["object_storage"].Description,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"service_name": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCPrivateEndpoint().Schema["service_name"].Description,
				Optional:    true,
			},
			"endpoint_address": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_dns_records_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"dns_records": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexVPCPrivateEndpointRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "private_endpoint_id", "name")
	if err != nil {
		return err
	}

	peID := d.Get("private_endpoint_id").(string)
	_, peNameOk := d.GetOk("name")

	if peNameOk {
		peID, err = resolvePrivateEndpointIDByName(ctx, config, d)
		if err != nil {
			return fmt.Errorf("failed to resolve data source private endpoint by name: %v", err)
		}
	}

	d.SetId(peID)

	if err := d.Set("private_endpoint_id", peID); err != nil {
		return err
	}

	return yandexVPCPrivateEndpointRead(d, meta, peID)
}

func resolvePrivateEndpointIDByName(ctx context.Context, config *Config, d *schema.ResourceData) (string, error) {
	name := d.Get("name").(string)
	folderID, err := getFolderID(d, config)
	if err != nil {
		return "", err
	}
	resolver := sdkresolversv2.NewBaseNameResolver(name, "PrivateEndpoint", sdkresolversv2.FolderID(folderID))
	client := privatelinksdk.NewPrivateEndpointClient(config.SDK)
	items, listErr := client.Iterator(ctx, &privatelink.ListPrivateEndpointsRequest{
		Container: &privatelink.ListPrivateEndpointsRequest_FolderId{FolderId: folderID},
		Filter:    sdkresolversv2.CreateResolverFilter("name", name),
		PageSize:  sdkresolversv2.DefaultResolverPageSize,
	}).TakeAll()
	if err := resolver.FindName(items, listErr); err != nil {
		return "", err
	}
	return resolver.ID(), nil
}
