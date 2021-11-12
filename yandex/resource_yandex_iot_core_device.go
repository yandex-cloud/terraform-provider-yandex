package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
)

func resourceYandexIoTCoreDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIoTCoreDeviceCreate,
		Read:   resourceYandexIoTCoreDeviceRead,
		Update: resourceYandexIoTCoreDeviceUpdate,
		Delete: resourceYandexIoTCoreDeviceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexIoTDefaultTimeout),
			Update: schema.DefaultTimeout(yandexIoTDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexIoTDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"certificates": {
				Type:     schema.TypeSet,
				MaxItems: 5,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"passwords": {
				Type:      schema.TypeSet,
				MaxItems:  5,
				Optional:  true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Set:       schema.HashString,
				Sensitive: true,
			},

			"aliases": {
				Type:     schema.TypeMap,
				Optional: true,
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

func resourceYandexIoTCoreDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	aliases, err := expandLabels(d.Get("aliases"))
	if err != nil {
		return fmt.Errorf("Error expanding aliases while creating IoT Device: %s", err)
	}

	certsSet := expandIoTCerts(d)
	var certs []*iot.CreateDeviceRequest_Certificate
	for cert := range certsSet {
		certs = append(certs, &iot.CreateDeviceRequest_Certificate{CertificateData: cert})
	}

	req := iot.CreateDeviceRequest{
		RegistryId:   d.Get("registry_id").(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Certificates: certs,
		TopicAliases: aliases,
	}

	op, err := config.sdk.WrapOperation(config.sdk.IoT().Devices().Device().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Device: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Device: %s", err)
	}

	md, ok := protoMetadata.(*iot.CreateDeviceMetadata)
	if !ok {
		return fmt.Errorf("Could not get IoT Device ID from create operation metadata")
	}

	d.SetId(md.DeviceId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Device: %s", err)
	}

	err = addDevicePasswords(ctx, config, d)
	if err != nil {
		return fmt.Errorf("Failed to set IoT Device password(s): %s", err)
	}

	return resourceYandexIoTCoreDeviceRead(d, meta)
}

func flattenYandexIoTCoreDevice(d *schema.ResourceData, device *iot.Device) error {
	d.Set("registry_id", device.RegistryId)
	d.Set("name", device.Name)
	d.Set("description", device.Description)
	d.Set("created_at", getTimestamp(device.CreatedAt))

	return nil
}

func resourceYandexIoTCoreDeviceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := iot.GetDeviceRequest{
		DeviceId: d.Id(),
	}

	device, err := config.sdk.IoT().Devices().Device().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Device %q", d.Id()))
	}

	return flattenYandexIoTCoreDevice(d, device)
}

func resourceYandexIoTCoreDeviceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := iot.DeleteDeviceRequest{
		DeviceId: d.Id(),
	}

	op, err := config.sdk.IoT().Devices().Device().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Device %q", d.Id()))
	}

	return nil
}

func resourceYandexIoTCoreDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	aliases, err := expandLabels(d.Get("aliases"))
	if err != nil {
		return fmt.Errorf("Error expanding aliases while updating IoT Device: %s", err)
	}

	d.Partial(true)

	var updatePaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
	}

	if d.HasChange("aliases") {
		updatePaths = append(updatePaths, "topic_aliases")
	}

	if len(updatePaths) != 0 {
		req := iot.UpdateDeviceRequest{
			DeviceId:     d.Id(),
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			TopicAliases: aliases,
			UpdateMask:   &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := config.sdk.IoT().Devices().Device().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update IoT Device: %s", err)
		}
	}

	if d.HasChange("certificates") {
		certsSetInner := expandIoTCerts(d)

		certsResp, err := config.sdk.IoT().Devices().Device().ListCertificates(ctx, &iot.ListDeviceCertificatesRequest{DeviceId: d.Id()})
		if err != nil {
			return err
		}

		for _, cert := range certsResp.Certificates {
			_, ok := certsSetInner[cert.CertificateData]
			if !ok {
				op, err := config.sdk.IoT().Devices().Device().DeleteCertificate(ctx, &iot.DeleteDeviceCertificateRequest{DeviceId: d.Id(), Fingerprint: cert.Fingerprint})
				err = waitOperation(ctx, config, op, err)
				if err != nil {
					return fmt.Errorf("Failed to remove certificate: %s, fingerpring: %s", err, cert.Fingerprint)
				}
			} else {
				delete(certsSetInner, cert.CertificateData)
			}
		}

		for cert := range certsSetInner {
			op, err := config.sdk.IoT().Devices().Device().AddCertificate(ctx, &iot.AddDeviceCertificateRequest{DeviceId: d.Id(), CertificateData: cert})
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Failed to add certificate: %s", err)
			}
		}

	}

	if d.HasChange("passwords") {
		passResp, err := config.sdk.IoT().Devices().Device().ListPasswords(ctx, &iot.ListDevicePasswordsRequest{DeviceId: d.Id()})
		if err != nil {
			return err
		}

		for _, pass := range passResp.Passwords {
			op, err := config.sdk.IoT().Devices().Device().DeletePassword(ctx, &iot.DeleteDevicePasswordRequest{DeviceId: d.Id(), PasswordId: pass.Id})
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Failed to delete password: %s", err)
			}
		}

		err = addDevicePasswords(ctx, config, d)
		if err != nil {
			return fmt.Errorf("Failed to add password: %s", err)
		}

	}

	d.Partial(false)

	return resourceYandexIoTCoreDeviceRead(d, meta)
}

func addDevicePasswords(ctx context.Context, config *Config, d *schema.ResourceData) error {
	passwordsSet := expandIoTPasswords(d)
	for pass := range passwordsSet {
		req := iot.AddDevicePasswordRequest{
			DeviceId: d.Id(),
			Password: pass,
		}

		op, err := config.sdk.IoT().Devices().Device().AddPassword(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return err
		}
	}
	return nil
}
