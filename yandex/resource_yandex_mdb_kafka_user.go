package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	kafkasdk "github.com/yandex-cloud/go-sdk/services/mdb/kafka/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBKafkaUserCreateTimeout = 10 * time.Minute
	yandexMDBKafkaUserReadTimeout   = 1 * time.Minute
	yandexMDBKafkaUserUpdateTimeout = 10 * time.Minute
	yandexMDBKafkaUserDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBKafkaUser() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a user of a Kafka User within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",

		Create: resourceYandexMDBKafkaUserCreate,
		Read:   resourceYandexMDBKafkaUserRead,
		Update: resourceYandexMDBKafkaUserUpdate,
		Delete: resourceYandexMDBKafkaUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			return validateKafkaUserPasswordConflict(d)
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBKafkaUserCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBKafkaUserReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBKafkaUserUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBKafkaUserDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Kafka cluster.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Type:         schema.TypeString,
				Description:  "The password of the user.",
				Optional:     true,
				Sensitive:    true,
				AtLeastOneOf: []string{"password", "password_wo"},
			},
			"password_wo": {
				Type:         schema.TypeString,
				Description:  "The password of the user. This attribute is write-only and is not stored in state. Requires `password_wo_version` to trigger updates. Write-only arguments are only supported in Terraform v1.11 or higher.",
				Optional:     true,
				WriteOnly:    true,
				Sensitive:    true,
				AtLeastOneOf: []string{"password", "password_wo"},
				RequiredWith: []string{"password_wo_version"},
			},
			"password_wo_version": {
				Type:         schema.TypeInt,
				Description:  "A version number for the write-only password. Increment this to trigger a password update.",
				Optional:     true,
				RequiredWith: []string{"password_wo"},
			},
			"permission": {
				Type:        schema.TypeSet,
				Description: "Set of permissions granted to the user.",
				Optional:    true,
				Set:         kafkaUserPermissionHash,
				Elem:        resourceYandexMDBKafkaPermission(),
			},
		},
	}
}

func validateKafkaUserPasswordConflict(d mdbcommon.RawConfigProvider) error {
	_, hasPassword := mdbcommon.LookupRawConfigPath(d, "password")
	_, hasPasswordWo := mdbcommon.LookupRawConfigPath(d, "password_wo")
	if hasPassword && hasPasswordWo {
		return fmt.Errorf("only one of `password` or `password_wo` can be specified")
	}
	return nil
}

func validateKafkaUserPasswordPair(d mdbcommon.RawConfigProvider) error {
	_, hasPasswordWo := mdbcommon.LookupRawConfigPath(d, "password_wo")
	_, hasPasswordWoVersion := mdbcommon.LookupRawConfigPath(d, "password_wo_version")
	if hasPasswordWo != hasPasswordWoVersion {
		return fmt.Errorf("`password_wo` and `password_wo_version` must be specified together")
	}
	return nil
}

type kafkaUserPasswordProvider interface {
	mdbcommon.RawConfigProvider
	GetOk(string) (any, bool)
}

func kafkaUserPassword(d kafkaUserPasswordProvider) string {
	if passwordWo, ok := mdbcommon.LookupRawConfigPath(d, "password_wo"); ok {
		return passwordWo.AsString()
	}
	if password, ok := d.GetOk("password"); ok {
		return password.(string)
	}
	return ""
}

func resourceYandexMDBKafkaUserCreate(d *schema.ResourceData, meta interface{}) error {
	if err := validateKafkaUserPasswordConflict(d); err != nil {
		return err
	}
	if err := validateKafkaUserPasswordPair(d); err != nil {
		return err
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	userSpec, err := buildKafkaUserSpec(d)
	if err != nil {
		return err
	}
	clusterID := d.Get("cluster_id").(string)
	if err = createKafkaUser(ctx, config, clusterID, userSpec); err != nil {
		return err
	}
	userID := constructResourceId(clusterID, userSpec.Name)
	d.SetId(userID)
	return resourceYandexMDBKafkaUserRead(d, meta)
}

func buildKafkaUserPermissions(d *schema.ResourceData) ([]*kafka.Permission, bool, error) {
	if permissionSchema, ok := d.GetOk("permission"); ok {
		permissions, err := expandKafkaPermissions(permissionSchema.(*schema.Set))
		if err != nil {
			return nil, false, err
		}
		return permissions, true, nil
	}
	return nil, false, nil
}

func buildKafkaUserSpec(d *schema.ResourceData) (*kafka.UserSpec, error) {
	userSpec := &kafka.UserSpec{
		Name:     d.Get("name").(string),
		Password: kafkaUserPassword(d),
	}
	permissions, ok, err := buildKafkaUserPermissions(d)
	if err != nil {
		return nil, err
	}
	if ok {
		userSpec.SetPermissions(permissions)
	}
	return userSpec, nil
}

func resourceYandexMDBKafkaUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()
	clusterID, userName, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}
	user, err := kafkasdk.NewUserClient(config.SDK).Get(ctx, &kafka.GetUserRequest{
		ClusterId: clusterID,
		UserName:  userName,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("User %q", userName))
	}
	perms := flattenKafkaUserPermissions(user)
	if err = d.Set("cluster_id", clusterID); err != nil {
		return err
	}
	if err = d.Set("name", user.Name); err != nil {
		return err
	}
	return d.Set("permission", perms)
}

func resourceYandexMDBKafkaUserUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := validateKafkaUserPasswordConflict(d); err != nil {
		return err
	}
	if err := validateKafkaUserPasswordPair(d); err != nil {
		return err
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	request := &kafka.UpdateUserRequest{
		ClusterId: d.Get("cluster_id").(string),
		UserName:  d.Get("name").(string),
		Password:  kafkaUserPassword(d),
	}

	permissions, ok, err := buildKafkaUserPermissions(d)
	if err != nil {
		return err
	}
	if ok {
		request.SetPermissions(permissions)
	}

	updatePaths := make([]string, 0, 2)
	for tfField, maskField := range mdbKafkaUserUpdateFieldsMap {
		if d.HasChange(tfField) {
			updatePaths = append(updatePaths, maskField)
		}
	}
	if len(updatePaths) == 0 {
		return nil
	}
	request.UpdateMask = &field_mask.FieldMask{Paths: updatePaths}

	if err = updateKafkaUser(ctx, config, request); err != nil {
		return err
	}
	return resourceYandexMDBKafkaUserRead(d, meta)
}

var mdbKafkaUserUpdateFieldsMap = map[string]string{
	"password":            "password",
	"password_wo_version": "password",
	"permission":          "permissions",
}

func resourceYandexMDBKafkaUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	userName := d.Get("name").(string)

	return deleteKafkaUser(ctx, config, clusterID, userName)
}
