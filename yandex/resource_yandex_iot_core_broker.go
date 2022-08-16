package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/broker/v1"
)

func resourceYandexIoTCoreBroker() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIoTCoreBrokerCreate,
		Read:   resourceYandexIoTCoreBrokerRead,
		Update: resourceYandexIoTCoreBrokerUpdate,
		Delete: resourceYandexIoTCoreBrokerDelete,

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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexIoTCoreBrokerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating IoT Broker: %s", err)
	}

	certsSet := expandIoTCerts(d)
	var certs []*iot.CreateBrokerRequest_Certificate
	for cert := range certsSet {
		certs = append(certs, &iot.CreateBrokerRequest_Certificate{CertificateData: cert})
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating IoT Broker: %s", err)
	}

	req := iot.CreateBrokerRequest{
		FolderId:     folderID,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		Certificates: certs,
	}

	op, err := config.sdk.WrapOperation(config.sdk.IoT().Broker().Broker().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Broker: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Broker: %s", err)
	}

	md, ok := protoMetadata.(*iot.CreateBrokerMetadata)
	if !ok {
		return fmt.Errorf("Could not get IoT Broker ID from create operation metadata")
	}

	d.SetId(md.BrokerId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Broker: %s", err)
	}

	return resourceYandexIoTCoreBrokerRead(d, meta)
}

func flattenYandexIoTCoreBroker(d *schema.ResourceData, broker *iot.Broker) error {
	d.Set("name", broker.Name)
	d.Set("description", broker.Description)
	d.Set("folder_id", broker.FolderId)
	if err := d.Set("labels", broker.Labels); err != nil {
		return err
	}
	d.Set("created_at", getTimestamp(broker.CreatedAt))
	return nil
}

func resourceYandexIoTCoreBrokerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := iot.GetBrokerRequest{
		BrokerId: d.Id(),
	}

	broker, err := config.sdk.IoT().Broker().Broker().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Broker %q", d.Id()))
	}

	return flattenYandexIoTCoreBroker(d, broker)
}

func resourceYandexIoTCoreBrokerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := iot.DeleteBrokerRequest{
		BrokerId: d.Id(),
	}

	op, err := config.sdk.IoT().Broker().Broker().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Broker %q", d.Id()))
	}

	return nil
}

func resourceYandexIoTCoreBrokerUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating IoT Broker: %s", err)
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

	if len(updatePaths) != 0 {
		req := iot.UpdateBrokerRequest{
			BrokerId:    d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := config.sdk.IoT().Broker().Broker().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update IoT Broker: %s", err)
		}

	}

	if d.HasChange("certificates") {
		certsSetInner := expandIoTCerts(d)

		certsResp, err := config.sdk.IoT().Broker().Broker().ListCertificates(ctx, &iot.ListBrokerCertificatesRequest{BrokerId: d.Id()})
		if err != nil {
			return err
		}

		for _, cert := range certsResp.Certificates {
			_, ok := certsSetInner[cert.CertificateData]
			if !ok {
				op, err := config.sdk.IoT().Broker().Broker().DeleteCertificate(ctx, &iot.DeleteBrokerCertificateRequest{BrokerId: d.Id(), Fingerprint: cert.Fingerprint})
				err = waitOperation(ctx, config, op, err)
				if err != nil {
					return fmt.Errorf("Failed to delete certificate: %s, fingerpring: %s", err, cert.Fingerprint)
				}
			} else {
				delete(certsSetInner, cert.CertificateData)
			}
		}

		for cert := range certsSetInner {
			op, err := config.sdk.IoT().Broker().Broker().AddCertificate(ctx, &iot.AddBrokerCertificateRequest{BrokerId: d.Id(), CertificateData: cert})
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Failed to add certificate: %s", err)
			}
		}

	}

	d.Partial(false)

	return resourceYandexIoTCoreBrokerRead(d, meta)
}
