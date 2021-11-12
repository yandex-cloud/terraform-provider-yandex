package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexIoTCoreDevice() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexIotCoreDeviceRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"passwords": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"aliases": {
				Type:     schema.TypeMap,
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

func dataSourceYandexIotCoreDeviceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	err := checkOneOf(d, "device_id", "name")
	if err != nil {
		return err
	}

	devID := d.Get("device_id").(string)
	_, ok := d.GetOk("name")

	if ok {
		devID, err = resolveObjectID(ctx, config, d, sdkresolvers.DeviceResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source IoT Device by name: %v", err)
		}
	}

	req := iot.GetDeviceRequest{
		DeviceId: devID,
	}

	device, err := config.sdk.IoT().Devices().Device().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Device %q", d.Id()))
	}

	certsResp, err := config.sdk.IoT().Devices().Device().ListCertificates(ctx, &iot.ListDeviceCertificatesRequest{DeviceId: devID})
	if err != nil {
		return err
	}

	var certs []string
	for _, cert := range certsResp.Certificates {
		certs = append(certs, cert.Fingerprint)
	}

	passResp, err := config.sdk.IoT().Devices().Device().ListPasswords(ctx, &iot.ListDevicePasswordsRequest{DeviceId: devID})
	if err != nil {
		return err
	}

	var passwords []string
	for _, pass := range passResp.Passwords {
		passwords = append(passwords, pass.Id)
	}

	d.SetId(device.Id)
	d.Set("device_id", device.Id)
	if err := flattenYandexIoTCoreDevice(d, device); err != nil {
		return err
	}
	if err := d.Set("aliases", device.TopicAliases); err != nil {
		return err
	}
	if err := d.Set("certificates", flattenIoTSet(certs)); err != nil {
		return err
	}
	return d.Set("passwords", flattenIoTSet(passwords))
}
