package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/broker/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexIoTCoreBroker() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex IoT Core Broker. For more information IoT Core, see [Yandex Cloud IoT Broker](https://yandex.cloud/docs/iot-core/quickstart).\nThis data source is used to define [Yandex Cloud IoT Broker](https://yandex.cloud/docs/iot-core/quickstart) that can be used by other resources.\n\n~> Either `broker_id` or `name` must be specified.\n",

		Read: dataSourceYandexIotCoreBrokerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"broker_id": {
				Type:        schema.TypeString,
				Description: "IoT Core Broker id used to define broker.",
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
				Description: resourceYandexIoTCoreBroker().Schema["certificates"].Description,
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

func dataSourceYandexIotCoreBrokerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	err := checkOneOf(d, "broker_id", "name")
	if err != nil {
		return err
	}

	brkID := d.Get("broker_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		brkID, err = resolveObjectID(ctx, config, d, sdkresolvers.BrokerResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source IoT Broker by name: %v", err)
		}
	}

	req := iot.GetBrokerRequest{
		BrokerId: brkID,
	}

	broker, err := config.sdk.IoT().Broker().Broker().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Broker %q", d.Id()))
	}

	certsResp, err := config.sdk.IoT().Broker().Broker().ListCertificates(ctx, &iot.ListBrokerCertificatesRequest{BrokerId: brkID})
	if err != nil {
		return err
	}

	var certs []string
	for _, cert := range certsResp.Certificates {
		certs = append(certs, cert.Fingerprint)
	}

	passResp, err := config.sdk.IoT().Broker().Broker().ListPasswords(ctx, &iot.ListBrokerPasswordsRequest{BrokerId: brkID})
	if err != nil {
		return err
	}

	// TODO: SA4010: this result of append is never used, except maybe in other appends (staticcheck)
	var passwords []string
	for _, pass := range passResp.Passwords {
		passwords = append(passwords, pass.Id)
	}

	d.SetId(broker.Id)
	d.Set("broker_id", broker.Id)
	if err := flattenYandexIoTCoreBroker(d, broker); err != nil {
		return err
	}
	return d.Set("certificates", flattenIoTSet(certs))
}
