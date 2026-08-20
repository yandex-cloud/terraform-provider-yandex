package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func resourceYandexIoTCoreDevice() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud IoT Device](https://yandex.cloud/docs/iot-core/quickstart).",

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
				Type:        schema.TypeString,
				Description: "IoT Core Registry ID for the IoT Core Device.",
				Required:    true,
				ForceNew:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"certificates": {
				Type:        schema.TypeSet,
				Description: "A set of certificate's fingerprints for the IoT Core Device.",
				MaxItems:    5,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"passwords": {
				Type:        schema.TypeSet,
				Description: "A set of passwords's id for the IoT Core Device.",
				MaxItems:    5,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Sensitive:   true,
			},

			"aliases": {
				Type:        schema.TypeMap,
				Description: "A set of key/value aliases pairs to assign to the IoT Core Device.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func resourceYandexIoTCoreDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewDeviceClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating IoT Device: %s", err)
	}

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
		Labels:       labels,
		Certificates: certs,
		TopicAliases: aliases,
	}

	op, err := client.Create(ctx, &req)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Device: %s", err)
	}

	d.SetId(op.Metadata().DeviceId)

	_, err = op.Wait(ctx)
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
	if err := d.Set("labels", device.Labels); err != nil {
		return err
	}

	return nil
}

func resourceYandexIoTCoreDeviceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewDeviceClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := iot.GetDeviceRequest{
		DeviceId: d.Id(),
	}

	device, err := client.Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Device %q", d.Id()))
	}

	return flattenYandexIoTCoreDevice(d, device)
}

func resourceYandexIoTCoreDeviceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewDeviceClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := iot.DeleteDeviceRequest{
		DeviceId: d.Id(),
	}

	op, err := client.Delete(ctx, &req)
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Device %q", d.Id()))
	}

	return nil
}

func resourceYandexIoTCoreDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewDeviceClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating IoT Registry: %s", err)
	}

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

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
	}

	if d.HasChange("aliases") {
		updatePaths = append(updatePaths, "topic_aliases")
	}

	if len(updatePaths) != 0 {
		req := iot.UpdateDeviceRequest{
			DeviceId:     d.Id(),
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			Labels:       labels,
			TopicAliases: aliases,
			UpdateMask:   &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := client.Update(ctx, &req)
		if err == nil {
			_, err = op.Wait(ctx)
		}
		if err != nil {
			return fmt.Errorf("Error while requesting API to update IoT Device: %s", err)
		}
	}

	if d.HasChange("certificates") {
		certsSetInner := expandIoTCerts(d)

		certsResp, err := client.ListCertificates(ctx, &iot.ListDeviceCertificatesRequest{DeviceId: d.Id()})
		if err != nil {
			return err
		}

		for _, cert := range certsResp.Certificates {
			_, ok := certsSetInner[cert.CertificateData]
			if !ok {
				op, err := client.DeleteCertificate(ctx, &iot.DeleteDeviceCertificateRequest{DeviceId: d.Id(), Fingerprint: cert.Fingerprint})
				if err == nil {
					_, err = op.Wait(ctx)
				}
				if err != nil {
					return fmt.Errorf("Failed to remove certificate: %s, fingerpring: %s", err, cert.Fingerprint)
				}
			} else {
				delete(certsSetInner, cert.CertificateData)
			}
		}

		for cert := range certsSetInner {
			op, err := client.AddCertificate(ctx, &iot.AddDeviceCertificateRequest{DeviceId: d.Id(), CertificateData: cert})
			if err == nil {
				_, err = op.Wait(ctx)
			}
			if err != nil {
				return fmt.Errorf("Failed to add certificate: %s", err)
			}
		}

	}

	if d.HasChange("passwords") {
		passResp, err := client.ListPasswords(ctx, &iot.ListDevicePasswordsRequest{DeviceId: d.Id()})
		if err != nil {
			return err
		}

		for _, pass := range passResp.Passwords {
			op, err := client.DeletePassword(ctx, &iot.DeleteDevicePasswordRequest{DeviceId: d.Id(), PasswordId: pass.Id})
			if err == nil {
				_, err = op.Wait(ctx)
			}
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
	client := devicessdk.NewDeviceClient(config.SDK)

	passwordsSet := expandIoTPasswords(d)
	for pass := range passwordsSet {
		req := iot.AddDevicePasswordRequest{
			DeviceId: d.Id(),
			Password: pass,
		}

		op, err := client.AddPassword(ctx, &req)
		if err == nil {
			_, err = op.Wait(ctx)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
