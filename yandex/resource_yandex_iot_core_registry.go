package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const yandexIoTDefaultTimeout = 5 * time.Minute

func resourceYandexIoTCoreRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIoTCoreRegistryCreate,
		Read:   resourceYandexIoTCoreRegistryRead,
		Update: resourceYandexIoTCoreRegistryUpdate,
		Delete: resourceYandexIoTCoreRegistryDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexIoTDefaultTimeout),
			Update: schema.DefaultTimeout(yandexIoTDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexIoTDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexIoTCoreRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating IoT Registry: %s", err)
	}

	certsSet := expandIoTCerts(d)
	var certs []*iot.CreateRegistryRequest_Certificate
	for cert := range certsSet {
		certs = append(certs, &iot.CreateRegistryRequest_Certificate{CertificateData: cert})
	}

	req := iot.CreateRegistryRequest{
		FolderId:     config.FolderID,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		Certificates: certs,
	}

	op, err := config.sdk.WrapOperation(config.sdk.IoT().Devices().Registry().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Registry: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Registry: %s", err)
	}

	md, ok := protoMetadata.(*iot.CreateRegistryMetadata)
	if !ok {
		return fmt.Errorf("Could not get IoT Registry ID from create operation metadata")
	}

	d.SetId(md.RegistryId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Registry: %s", err)
	}

	err = addRegistryPasswords(ctx, config, d)
	if err != nil {
		return fmt.Errorf("Failed to set IoT Registry password(s): %s", err)
	}

	return resourceYandexIoTCoreRegistryRead(d, meta)
}

func flattenYandexIoTCoreRegistry(d *schema.ResourceData, registry *iot.Registry) error {
	createdAt, err := getTimestamp(registry.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("name", registry.Name)
	d.Set("description", registry.Description)
	d.Set("folder_id", registry.FolderId)
	if err := d.Set("labels", registry.Labels); err != nil {
		return err
	}
	d.Set("created_at", createdAt)
	return nil
}

func resourceYandexIoTCoreRegistryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := iot.GetRegistryRequest{
		RegistryId: d.Id(),
	}

	registry, err := config.sdk.IoT().Devices().Registry().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Registry %q", d.Id()))
	}

	return flattenYandexIoTCoreRegistry(d, registry)
}

func resourceYandexIoTCoreRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := iot.DeleteRegistryRequest{
		RegistryId: d.Id(),
	}

	op, err := config.sdk.IoT().Devices().Registry().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Registry %q", d.Id()))
	}

	return nil
}

func resourceYandexIoTCoreRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating IoT Registry: %s", err)
	}

	d.Partial(true)

	var updatePaths []string
	var partialPaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
		partialPaths = append(partialPaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
		partialPaths = append(partialPaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
		partialPaths = append(partialPaths, "labels")
	}

	if len(updatePaths) != 0 {
		req := iot.UpdateRegistryRequest{
			RegistryId:  d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := config.sdk.IoT().Devices().Registry().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update IoT Registry: %s", err)
		}

		for _, v := range partialPaths {
			d.SetPartial(v)
		}
	}

	if d.HasChange("certificates") {
		certsSetInner := expandIoTCerts(d)

		certsResp, err := config.sdk.IoT().Devices().Registry().ListCertificates(ctx, &iot.ListRegistryCertificatesRequest{RegistryId: d.Id()})
		if err != nil {
			return err
		}

		for _, cert := range certsResp.Certificates {
			_, ok := certsSetInner[cert.CertificateData]
			if !ok {
				op, err := config.sdk.IoT().Devices().Registry().DeleteCertificate(ctx, &iot.DeleteRegistryCertificateRequest{RegistryId: d.Id(), Fingerprint: cert.Fingerprint})
				err = waitOperation(ctx, config, op, err)
				if err != nil {
					return fmt.Errorf("Failed to delete certificate: %s, fingerpring: %s", err, cert.Fingerprint)
				}
			} else {
				delete(certsSetInner, cert.CertificateData)
			}
		}

		for cert := range certsSetInner {
			op, err := config.sdk.IoT().Devices().Registry().AddCertificate(ctx, &iot.AddRegistryCertificateRequest{RegistryId: d.Id(), CertificateData: cert})
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Failed to add certificate: %s", err)
			}
		}

		d.SetPartial("certificates")
	}

	if d.HasChange("passwords") {
		passResp, err := config.sdk.IoT().Devices().Registry().ListPasswords(ctx, &iot.ListRegistryPasswordsRequest{RegistryId: d.Id()})
		if err != nil {
			return err
		}
		passwordsSet := expandIoTPasswords(d)

		if len(passResp.Passwords) == len(passwordsSet) {
			err = addRegistryPasswords(ctx, config, d)
			if err != nil {
				return fmt.Errorf("Failed to add password: %s", err)
			}
		} else {
			for _, pass := range passResp.Passwords {
				op, err := config.sdk.IoT().Devices().Registry().DeletePassword(ctx, &iot.DeleteRegistryPasswordRequest{RegistryId: d.Id(), PasswordId: pass.Id})
				err = waitOperation(ctx, config, op, err)
				if err != nil {
					return fmt.Errorf("Failed to delete password: %s", err)
				}
			}

			err = addRegistryPasswords(ctx, config, d)
			if err != nil {
				return fmt.Errorf("Failed to add password: %s", err)
			}
		}

		d.SetPartial("passwords")
	}

	d.Partial(false)

	return resourceYandexIoTCoreRegistryRead(d, meta)
}

func addRegistryPasswords(ctx context.Context, config *Config, d *schema.ResourceData) error {
	passwordsSet := expandIoTPasswords(d)
	for pass := range passwordsSet {
		req := iot.AddRegistryPasswordRequest{
			RegistryId: d.Id(),
			Password:   pass,
		}

		op, err := config.sdk.IoT().Devices().Registry().AddPassword(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return err
		}
	}
	return nil
}

func expandIoTSet(name string, d *schema.ResourceData) map[string]interface{} {
	result := make(map[string]interface{})
	set := d.Get(name).(*schema.Set)

	for _, t := range set.List() {
		cert := t.(string)
		result[cert] = nil
	}

	return result
}

func expandIoTCerts(d *schema.ResourceData) map[string]interface{} {
	return expandIoTSet("certificates", d)
}

func expandIoTPasswords(d *schema.ResourceData) map[string]interface{} {
	return expandIoTSet("passwords", d)
}

func waitOperation(ctx context.Context, config *Config, opInput *operation.Operation, err error) error {
	op, err := config.sdk.WrapOperation(opInput, err)
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}
