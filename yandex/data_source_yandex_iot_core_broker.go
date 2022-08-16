package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/broker/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexIoTCoreBroker() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexIotCoreBrokerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"broker_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"certificates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
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
