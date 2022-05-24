package yandex

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBMySQLUserCreateTimeout = 10 * time.Minute
	yandexMDBMySQLUserReadTimeout   = 1 * time.Minute
	yandexMDBMySQLUserUpdateTimeout = 10 * time.Minute
	yandexMDBMySQLUserDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBMySQLUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBMySQLUserCreate,
		Read:   resourceYandexMDBMySQLUserRead,
		Update: resourceYandexMDBMySQLUserUpdate,
		Delete: resourceYandexMDBMySQLUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMySQLUserCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBMySQLUserReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBMySQLUserUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMySQLUserDeleteTimeout),
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
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      mysqlUserPermissionHash,
				Elem:     resourceYandexMDBMySQLUserPermission(),
			},
			"global_permissions": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},
			"connection_limits": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBMySQLUserConnectionLimits(),
			},
			"authentication_plugin": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexMDBMySQLUserPermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceYandexMDBMySQLUserConnectionLimits() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_questions_per_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"max_updates_per_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"max_connections_per_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"max_user_connections": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
		},
	}
}

func resourceYandexMDBMySQLUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	userSpec, err := expandMySQLUserSpec(d)
	if err != nil {
		return err
	}
	request := &mysql.CreateUserRequest{
		ClusterId: clusterID,
		UserSpec:  userSpec,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MySQL user create request: %+v", request)
		return config.sdk.MDB().MySQL().User().Create(ctx, request)
	})

	userID := constructResourceId(clusterID, userSpec.Name)
	d.SetId(userID)

	if err != nil {
		return fmt.Errorf("error while requesting API to create user for MySQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while creating user for MySQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating user for MySQL Cluster %q failed: %s", clusterID, err)
	}

	return resourceYandexMDBMySQLUserRead(d, meta)
}

func expandMySQLUserSpec(d *schema.ResourceData) (*mysql.UserSpec, error) {
	user := &mysql.UserSpec{}

	if v, ok := d.GetOk("name"); ok {
		user.Name = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		user.Password = v.(string)
	}

	if v, ok := d.GetOk("permission"); ok {
		permissions, err := expandMysqlUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	if v, ok := d.GetOk("global_permissions"); ok {
		gs, err := expandMysqlUserGlobalPermissions(v.(*schema.Set).List())
		if err != nil {
			return nil, err
		}
		user.GlobalPermissions = gs
	}

	if conLimits, ok := d.GetOk("connection_limits"); ok {
		connectionLimitsMap := (conLimits.([]interface{}))[0].(map[string]interface{})
		user.ConnectionLimits = expandMySQLConnectionLimits(connectionLimitsMap)
	}

	if v, ok := d.GetOk("authentication_plugin"); ok {
		authenticationPlugin, err := expandEnum("authentication_plugin", v.(string), mysql.AuthPlugin_value)
		if err != nil {
			return nil, err
		}
		user.AuthenticationPlugin = mysql.AuthPlugin(*authenticationPlugin)
	}

	return user, nil
}

func resourceYandexMDBMySQLUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	clusterID, username, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}

	user, err := config.sdk.MDB().MySQL().User().Get(ctx, &mysql.GetUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("User %q", username))
	}

	permissions, err := flattenMysqlUserPermissions(user.Permissions)
	if err != nil {
		return err
	}

	connectionLimits := flattenMysqlUserConnectionLimits(user)
	globalPermissions := unbindGlobalPermissions(user.GlobalPermissions)

	d.Set("cluster_id", clusterID)
	d.Set("name", user.Name)
	d.Set("permission", permissions)
	d.Set("global_permissions", globalPermissions)
	d.Set("connection_limits", connectionLimits)
	if user.AuthenticationPlugin != 0 {
		d.Set("authentication_plugin", mysql.AuthPlugin_name[int32(user.AuthenticationPlugin)])
	}
	return nil
}

func resourceYandexMDBMySQLUserUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	user, err := expandMySQLUserSpec(d)
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	request := &mysql.UpdateUserRequest{
		ClusterId:            clusterID,
		UserName:             user.Name,
		Password:             user.Password,
		Permissions:          user.Permissions,
		AuthenticationPlugin: user.AuthenticationPlugin,
		ConnectionLimits:     user.ConnectionLimits,
		GlobalPermissions:    user.GlobalPermissions,
		UpdateMask:           &field_mask.FieldMask{Paths: []string{"authentication_plugin", "password", "permissions", "connection_limits", "global_permissions"}},
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MySQL user update request: %+v", request)
		return config.sdk.MDB().MySQL().User().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update user in MySQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while updating user in MySQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating user for MySQL Cluster %q failed: %s", clusterID, err)
	}
	return nil
}

func resourceYandexMDBMySQLUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)

	request := &mysql.DeleteUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MySQL user delete request: %+v", request)
		return config.sdk.MDB().MySQL().User().Delete(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from MySQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while deleting user from MySQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting user from MySQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}
