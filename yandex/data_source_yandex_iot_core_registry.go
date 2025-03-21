package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexIoTCoreRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex IoT Core Registry. For more information IoT Core, see [Yandex Cloud IoT Registry](https://yandex.cloud/docs/iot-core/quickstart).\nThis data source is used to define [Yandex Cloud IoT Registry](https://yandex.cloud/docs/iot-core/quickstart) that can be used by other resources.\n\n~> Either `registry_id` or `name` must be specified.\n",

		Read: dataSourceYandexIotCoreRegistryRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:        schema.TypeString,
				Description: "IoT Core Registry id used to define registry.",
				Optional:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
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

			"certificates": {
				Type:        schema.TypeSet,
				Description: resourceYandexIoTCoreRegistry().Schema["certificates"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"passwords": {
				Type:        schema.TypeSet,
				Description: resourceYandexIoTCoreRegistry().Schema["passwords"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"log_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"log_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"min_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func flattenIoTSet(vars []string) *schema.Set {
	result := &schema.Set{F: schema.HashString}
	for _, v := range vars {
		result.Add(v)
	}
	return result
}

func dataSourceYandexIotCoreRegistryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	err := checkOneOf(d, "registry_id", "name")
	if err != nil {
		return err
	}

	regID := d.Get("registry_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		regID, err = resolveObjectID(ctx, config, d, sdkresolvers.DeviceRegistryResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source IoT Registry by name: %v", err)
		}
	}

	req := iot.GetRegistryRequest{
		RegistryId: regID,
	}

	registry, err := config.sdk.IoT().Devices().Registry().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Registry %q", d.Id()))
	}

	certsResp, err := config.sdk.IoT().Devices().Registry().ListCertificates(ctx, &iot.ListRegistryCertificatesRequest{RegistryId: regID})
	if err != nil {
		return err
	}

	var certs []string
	for _, cert := range certsResp.Certificates {
		certs = append(certs, cert.Fingerprint)
	}

	passResp, err := config.sdk.IoT().Devices().Registry().ListPasswords(ctx, &iot.ListRegistryPasswordsRequest{RegistryId: regID})
	if err != nil {
		return err
	}

	var passwords []string
	for _, pass := range passResp.Passwords {
		passwords = append(passwords, pass.Id)
	}

	d.SetId(registry.Id)
	d.Set("registry_id", registry.Id)
	if err := flattenYandexIoTCoreRegistry(d, registry); err != nil {
		return err
	}
	if err := d.Set("certificates", flattenIoTSet(certs)); err != nil {
		return err
	}
	return d.Set("passwords", flattenIoTSet(passwords))
}
