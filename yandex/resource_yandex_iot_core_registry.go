package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
)

const yandexIoTDefaultTimeout = 5 * time.Minute

func resourceYandexIoTCoreRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud IoT Registry](https://yandex.cloud/docs/iot-core/quickstart).",

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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
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
				Description: "A set of certificate's fingerprints for the IoT Core Registry.",
				MaxItems:    5,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"passwords": {
				Type:        schema.TypeSet,
				Description: "A set of passwords's id for the IoT Core Registry.",
				MaxItems:    5,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Sensitive:   true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"log_options": {
				Type:        schema.TypeList,
				Description: "Options for logging for IoT Core Registry.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:        schema.TypeBool,
							Description: "Is logging for registry disabled.",
							Optional:    true,
						},
						"log_group_id": {
							Type:          schema.TypeString,
							Description:   "Log entries are written to specified log group.",
							Optional:      true,
							ConflictsWith: []string{"log_options.0.folder_id"},
							ExactlyOneOf:  []string{"log_options.0.folder_id", "log_options.0.log_group_id"},
						},
						"folder_id": {
							Type:          schema.TypeString,
							Description:   "Log entries are written to default log group for specified folder.",
							Optional:      true,
							ConflictsWith: []string{"log_options.0.log_group_id"},
							ExactlyOneOf:  []string{"log_options.0.folder_id", "log_options.0.log_group_id"},
						},
						"min_level": {
							Type:        schema.TypeString,
							Description: "Minimum log entry level.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexIoTCoreRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewRegistryClient(config.SDK)

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

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating IoT Registry: %s", err)
	}

	logOptions, err := expandRegistryLogOptions(d)
	if err != nil {
		return fmt.Errorf("Error expanding log options while creating IoT Registry: %s", err)
	}

	req := iot.CreateRegistryRequest{
		FolderId:     folderID,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		Certificates: certs,
		LogOptions:   logOptions,
	}

	op, err := client.Create(ctx, &req)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create IoT Registry: %s", err)
	}

	d.SetId(op.Metadata().RegistryId)

	_, err = op.Wait(ctx)
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
	d.Set("name", registry.Name)
	d.Set("description", registry.Description)
	d.Set("folder_id", registry.FolderId)
	if err := d.Set("labels", registry.Labels); err != nil {
		return err
	}
	d.Set("created_at", getTimestamp(registry.CreatedAt))
	if logOptions := flattenRegistryLogOptions(registry.LogOptions); logOptions != nil {
		d.Set("log_options", logOptions)
	}
	return nil
}

func resourceYandexIoTCoreRegistryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewRegistryClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := iot.GetRegistryRequest{
		RegistryId: d.Id(),
	}

	registry, err := client.Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Registry %q", d.Id()))
	}

	return flattenYandexIoTCoreRegistry(d, registry)
}

func resourceYandexIoTCoreRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewRegistryClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := iot.DeleteRegistryRequest{
		RegistryId: d.Id(),
	}

	op, err := client.Delete(ctx, &req)
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IoT Registry %q", d.Id()))
	}

	return nil
}

func resourceYandexIoTCoreRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := devicessdk.NewRegistryClient(config.SDK)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating IoT Registry: %s", err)
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

	if d.HasChange("log_options") {
		updatePaths = append(updatePaths, "log_options")
	}

	if len(updatePaths) != 0 {
		req := iot.UpdateRegistryRequest{
			RegistryId:  d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
		}

		req.LogOptions, err = expandRegistryLogOptions(d)
		if err != nil {
			return fmt.Errorf("Error expanding log options while updating IoT Registry: %s", err)
		}

		op, err := client.Update(ctx, &req)
		if err == nil {
			_, err = op.Wait(ctx)
		}
		if err != nil {
			return fmt.Errorf("Error while requesting API to update IoT Registry: %s", err)
		}

	}

	if d.HasChange("certificates") {
		certsSetInner := expandIoTCerts(d)

		certsResp, err := client.ListCertificates(ctx, &iot.ListRegistryCertificatesRequest{RegistryId: d.Id()})
		if err != nil {
			return err
		}

		for _, cert := range certsResp.Certificates {
			_, ok := certsSetInner[cert.CertificateData]
			if !ok {
				op, err := client.DeleteCertificate(ctx, &iot.DeleteRegistryCertificateRequest{RegistryId: d.Id(), Fingerprint: cert.Fingerprint})
				if err == nil {
					_, err = op.Wait(ctx)
				}
				if err != nil {
					return fmt.Errorf("Failed to delete certificate: %s, fingerpring: %s", err, cert.Fingerprint)
				}
			} else {
				delete(certsSetInner, cert.CertificateData)
			}
		}

		for cert := range certsSetInner {
			op, err := client.AddCertificate(ctx, &iot.AddRegistryCertificateRequest{RegistryId: d.Id(), CertificateData: cert})
			if err == nil {
				_, err = op.Wait(ctx)
			}
			if err != nil {
				return fmt.Errorf("Failed to add certificate: %s", err)
			}
		}

	}

	if d.HasChange("passwords") {
		passResp, err := client.ListPasswords(ctx, &iot.ListRegistryPasswordsRequest{RegistryId: d.Id()})
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
				op, err := client.DeletePassword(ctx, &iot.DeleteRegistryPasswordRequest{RegistryId: d.Id(), PasswordId: pass.Id})
				if err == nil {
					_, err = op.Wait(ctx)
				}
				if err != nil {
					return fmt.Errorf("Failed to delete password: %s", err)
				}
			}

			err = addRegistryPasswords(ctx, config, d)
			if err != nil {
				return fmt.Errorf("Failed to add password: %s", err)
			}
		}

	}

	d.Partial(false)

	return resourceYandexIoTCoreRegistryRead(d, meta)
}

func addRegistryPasswords(ctx context.Context, config *Config, d *schema.ResourceData) error {
	client := devicessdk.NewRegistryClient(config.SDK)

	passwordsSet := expandIoTPasswords(d)
	for pass := range passwordsSet {
		req := iot.AddRegistryPasswordRequest{
			RegistryId: d.Id(),
			Password:   pass,
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

func expandRegistryLogOptions(d *schema.ResourceData) (*iot.LogOptions, error) {
	if v, ok := d.GetOk("log_options.0"); ok {
		logOptionsMap := v.(map[string]interface{})
		logOptions := &iot.LogOptions{}

		if disabled, ok := logOptionsMap["disabled"]; ok {
			logOptions.Disabled = disabled.(bool)
		}
		if folderID, ok := logOptionsMap["folder_id"]; ok {
			logOptions.SetFolderId(folderID.(string))
		}
		if logGroupID, ok := logOptionsMap["log_group_id"]; ok {
			logOptions.SetLogGroupId(logGroupID.(string))
		}
		if level, ok := logOptionsMap["min_level"]; ok {
			if v, ok := logging.LogLevel_Level_value[level.(string)]; ok {
				logOptions.MinLevel = logging.LogLevel_Level(v)
			} else {
				return nil, fmt.Errorf("unknown log level: %s", level)
			}
		}
		return logOptions, nil
	}
	return nil, nil
}

func flattenRegistryLogOptions(logOptions *iot.LogOptions) []interface{} {
	if logOptions == nil {
		return nil
	}
	res := map[string]interface{}{
		"disabled":  logOptions.Disabled,
		"min_level": logging.LogLevel_Level_name[int32(logOptions.MinLevel)],
	}
	if logOptions.Destination != nil {
		switch d := logOptions.Destination.(type) {
		case *iot.LogOptions_LogGroupId:
			res["log_group_id"] = d.LogGroupId
		case *iot.LogOptions_FolderId:
			res["folder_id"] = d.FolderId
		}
	}
	return []interface{}{res}
}
