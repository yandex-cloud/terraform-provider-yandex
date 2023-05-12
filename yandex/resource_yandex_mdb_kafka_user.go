package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
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
		Create: resourceYandexMDBKafkaUserCreate,
		Read:   resourceYandexMDBKafkaUserRead,
		Update: resourceYandexMDBKafkaUserUpdate,
		Delete: resourceYandexMDBKafkaUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      kafkaUserPermissionHash,
				Elem:     resourceYandexMDBKafkaPermission(),
			},
		},
	}
}

func resourceYandexMDBKafkaUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	userSpec, err := buildKafkaUserSpec(d)
	if err != nil {
		return err
	}
	clusterID := d.Get("cluster_id").(string)
	// Need to set resource id to value of cluster id, cause this value used in createKafkaUser func
	d.SetId(clusterID)
	if err = createKafkaUser(ctx, config, d, userSpec); err != nil {
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
		Password: d.Get("password").(string),
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
	user, err := config.sdk.MDB().Kafka().User().Get(ctx, &kafka.GetUserRequest{
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
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	request := &kafka.UpdateUserRequest{
		ClusterId: d.Get("cluster_id").(string),
		UserName:  d.Get("name").(string),
		Password:  d.Get("password").(string),
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
	"password":   "password",
	"permission": "permissions",
}

func resourceYandexMDBKafkaUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	userName := d.Get("name").(string)
	// Need to set resource id to value of cluster id, cause this value used in deleteKafkaUser func
	d.SetId(clusterID)
	return deleteKafkaUser(ctx, config, d, userName)
}
