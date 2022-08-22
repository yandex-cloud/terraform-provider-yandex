package yandex

import (
	"fmt"
	"log"
	"time"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const (
	yandexMDBPostgreSQLUserCreateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLUserReadTimeout   = 1 * time.Minute
	yandexMDBPostgreSQLUserUpdateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLUserDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBPostgreSQLUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBPostgreSQLUserCreate,
		Read:   resourceYandexMDBPostgreSQLUserRead,
		Update: resourceYandexMDBPostgreSQLUserUpdate,
		Delete: resourceYandexMDBPostgreSQLUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBPostgreSQLUserCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBPostgreSQLUserReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBPostgreSQLUserUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBPostgreSQLUserDeleteTimeout),
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
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"grants": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// TODO change to permissions
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"conn_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"settings": {
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceYandexMDBPostgreSQLUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	userSpec, err := expandPgUserSpec(d)
	if err != nil {
		return err
	}
	request := &postgresql.CreateUserRequest{
		ClusterId: clusterID,
		UserSpec:  userSpec,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user create request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Create(ctx, request)
	})

	userID := constructResourceId(clusterID, userSpec.Name)
	d.SetId(userID)

	if err != nil {
		return fmt.Errorf("error while requesting API to create user for PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while creating user for PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating user for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return resourceYandexMDBPostgreSQLUserRead(d, meta)
}

func expandPgUserSpec(d *schema.ResourceData) (*postgresql.UserSpec, error) {
	user := &postgresql.UserSpec{}
	if v, ok := d.GetOkExists("name"); ok {
		user.Name = v.(string)
	}

	if v, ok := d.GetOkExists("password"); ok {
		user.Password = v.(string)
	}

	if v, ok := d.GetOkExists("login"); ok {
		user.Login = &wrappers.BoolValue{Value: v.(bool)}
	}

	if v, ok := d.GetOkExists("conn_limit"); ok {
		user.ConnLimit = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOkExists("permission"); ok {
		permissions, err := expandPGUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	if v, ok := d.GetOkExists("grants"); ok {
		gs, err := expandPGUserGrants(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		user.Grants = gs
	}

	if _, ok := d.GetOkExists("settings"); ok {
		if user.Settings == nil {
			user.Settings = &postgresql.UserSettings{}
		}

		err := expandResourceGenerate(mdbPGUserSettingsFieldsInfo, d, user.Settings, "settings.", true)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func resourceYandexMDBPostgreSQLUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	clusterID, username, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}

	user, err := config.sdk.MDB().PostgreSQL().User().Get(ctx, &postgresql.GetUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("User %q", username))
	}

	d.Set("cluster_id", clusterID)
	d.Set("name", user.Name)
	d.Set("login", user.Login.GetValue())
	d.Set("grants", user.Grants)
	d.Set("permission", flattenPGUserPermissions(user.Permissions))
	d.Set("conn_limit", user.ConnLimit)
	knownDefault := map[string]struct{}{
		"log_min_duration_statement": {},
	}
	settings, err := flattenResourceGenerateMapS(user.Settings, false, mdbPGUserSettingsFieldsInfo, false, true, knownDefault)
	if err != nil {
		return err
	}
	d.Set("settings", settings)

	return nil
}

func resourceYandexMDBPostgreSQLUserUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	user, err := expandPgUserSpec(d)
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	request := &postgresql.UpdateUserRequest{
		ClusterId:   clusterID,
		UserName:    user.Name,
		Password:    user.Password,
		Permissions: user.Permissions,
		ConnLimit:   user.ConnLimit.GetValue(),
		Login:       user.Login,
		Grants:      user.Grants,
		Settings:    user.Settings,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user update request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while updating user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating user for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}
	return nil
}

func resourceYandexMDBPostgreSQLUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)

	request := &postgresql.DeleteUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user delete request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Delete(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while deleting user from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting user from PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}
